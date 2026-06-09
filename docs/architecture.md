# Architecture

Runlet has three moving parts in the MVP:

```
┌──────────────────────┐        HTTP / JSON         ┌──────────────────────┐
│  Core (Rails app)    │ ◀──────────────────────────│  Runner (Go binary)  │
│                      │                            │                      │
│  • Identities        │  POST /runners             │  • register          │
│  • Accounts          │  PATCH /runners/:id        │  • heartbeat         │
│  • Runners           │  POST /runners/:id/claims  │  • claim             │
│  • RunnerRuns        │  POST /runs/:id/events     │  • execute (sh -c)   │
│  • RunnerRunEvents   │  POST /runs/:id/finish     │  • stream events     │
└──────────────────────┘                            └──────────────────────┘
          ▲                                                    │
          │  Web UI (Turbo/Stimulus)                           │  os.exec
          │                                                    ▼
       Operators                                          Local shell
```

## Components

### Core (`core/`)

Rails 8 application using SQLite + Solid Queue / Cache / Cable. Owns:

- **Identities & accounts** — multi-tenant ownership via the `runlet.account_slug` middleware. The first non-reserved URL segment is the account slug and gets moved into `SCRIPT_NAME`.
- **Sessions** — magic-link based browser sessions and signed runner tokens. The runner uses a `Session#signed_id` as its bearer token.
- **Runners** — one row per registered worker (`account_id`, `name` is the unique scope; identity, labels, last heartbeat, current run).
- **RunnerRuns** — queued/running/finished shell commands targeted at a specific runner.
- **RunnerRunEvents** — append-only event log for each run, with sequence numbers unique per run.

### Runner (`apps/runner/`)

Small Go binary structured as four internal packages:

- `internal/config` — layered configuration (defaults → `~/.runlet/settings.json` → environment → flags) plus interactive `setup` prompts.
- `internal/daemon` — background runner process started by `setup`.
- `internal/api` — typed HTTP client for the Core API (`RegisterRunner`, `Heartbeat`, `Claim`, `SendRunEvent`, `FinishRun`).
- `internal/executor` — runs a single shell command in a workspace, streams stdout/stderr line-by-line, enforces timeouts, cleans up the process group.
- `internal/runner` — service that ties registration, heartbeats, polling, and execution together. Implements the `Once` mode used by tests and CI.

### Shared packages (`packages/`)

- `packages/types` — shared TypeScript payload shapes (currently a single re-export to be filled in as the JS client surface grows).
- `packages/ui` — design tokens and base CSS used by Core's web views.
- `packages/brand` — logos, brand materials, and static previews.

## Data model

```
identities ─┬─< sessions
            └─< users >─┬─ accounts
                        ├─< runners >─< runner_runs >─< runner_run_events
                        └─< device_authorizations
```

UUIDv7 primary keys. `runner_runs` and `runners` are always scoped to an `account_id`; the API authentication layer pins them to the slug carried by the URL.

## Request lifecycle

1. Browser / runner hits `/<account-slug>/api/v1/...`.
2. `AccountSlug::Extractor` middleware moves the slug to `SCRIPT_NAME` and sets `Current.account`.
3. `Api::V1::BaseController` runs `ApiAuthentication`: parse `Authorization: Bearer …` (session signed id or identity access token), set `Current.identity` / `Current.session`, then verify the identity is a user of `Current.account`.
4. Concrete controller (`Runners`, `Runs::Events`, `Runs::Finishes`, …) executes the action.
5. `ActiveRecord::RecordNotFound` is rendered as 404 JSON.

`Api::V1::BaseController` also enforces a minimum desktop client version (currently 1.6.0). The runner does not advertise itself as a desktop client, so the check is a no-op for it.

## Concurrency model

- The runner enforces `concurrency = 1` for the MVP; the configuration explicitly rejects other values.
- `RunnerRun#claim!` takes a row lock, verifies the run is still `queued`, then moves it to `running`. Two runners racing on the same row will see only one win.
- The Go executor scans `stdout` and `stderr` in dedicated goroutines and **waits for both to drain to EOF** before calling `cmd.Wait()`. `cmd.Wait()` closes the pipes on return, so draining first is what guarantees we don't lose the trailing lines of output. See [`troubleshooting.md`](troubleshooting.md#dropped-output-events-fixed) for the underlying bug history.
