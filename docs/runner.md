# Runner

`apps/runner` is the Go worker that picks up shell commands from Core and runs them locally. This is the only client surface in the MVP — everything else is queued from the Rails web UI or `bin/rails runner`.

## Install / build

```bash
cd apps/runner
go build -o ./bin/runlet-runner ./cmd/runlet-runner
```

Or, from the repo root, `pnpm runner:build`.

## Configure

`runlet-runner setup` writes a key/value config file to the OS user config directory (e.g. `~/Library/Application Support/runlet/runner.conf` on macOS, mode `0600`).

```bash
runlet-runner setup <session-token> \
  --api-url http://localhost:3000/<account-slug> \
  --workspace /Users/me/Projects/runlet \
  --label kind=desktop --label project=runlet
```

`<session-token>` is a Rails `Session#signed_id`. Generate one from a console:

```bash
cd core
bin/rails runner '
identity = Identity.find_by!(email: "you@example.com")
session = identity.sessions.create!(user_agent: "Runlet CLI", ip_address: "127.0.0.1")
puts session.signed_id
'
```

> Tip: the account slug **must** appear in `--api-url`; the runner does not append it for you.

`setup` always clears any previously-saved `RUNLET_RUNNER_ID` so the first run re-registers under the configured name.

### Layered configuration

Each runtime invocation merges configuration in this order, with later sources overriding earlier ones:

1. Hard-coded defaults (`http://localhost:3000`, `dev-token`, name `local-runner` overridden by hostname during `applyDefaults`, current directory as workspace, labels `kind=desktop,project=runlet` plus auto-injected `os` and `arch`).
2. The on-disk `runner.conf` written by `setup`.
3. `RUNLET_*` environment variables.
4. Command-line flags (`-api-url`, `-token`, `-workspace`, repeated `-label key=value`, etc.).
5. If stdin is a terminal and `-non-interactive` is not set, the runner prompts for each value with the merged result as the default.

Supported environment variables:

```
RUNLET_API_URL                       e.g. http://localhost:3000/acme
RUNLET_TOKEN                         session signed id or access token
RUNLET_RUNNER_ID                     skip registration and reuse this id
RUNLET_RUNNER_NAME                   defaults to hostname (or `runlet-runner` when no hostname)
RUNLET_WORKSPACE                     path used when a run cwd is empty or relative (no absolute-path enforcement; caller is responsible)
RUNLET_SHELL                         override shell (defaults to /bin/sh; cmd.exe on Windows)
RUNLET_LABELS                        kind=desktop,project=runlet,...
RUNLET_CONCURRENCY                   must be 1 in the MVP
RUNLET_POLL_INTERVAL_SECONDS         default 5
RUNLET_HEARTBEAT_INTERVAL_SECONDS    default 15
RUNLET_DEFAULT_TIMEOUT_SECONDS       default 900 (15 minutes)
```

## Run

```bash
runlet-runner                  # poll forever
runlet-runner -once            # claim and execute at most one run, then exit
runlet-runner -non-interactive # never prompt; required values must come from flags/env
runlet-runner -config /tmp/test/runner.conf
```

The runner logs to stdout using `log/slog`'s text handler. Typical output:

```
time=… level=INFO msg="registering runner" name=local-runner
time=… level=INFO msg="runner registered" runner_id=…
time=… level=INFO msg="executing run" run_id=… cwd=/workspace
time=… level=INFO msg="run finished" run_id=… status=succeeded
```

## Execution semantics

- Each claimed run is executed by spawning `/bin/sh -c "<command>"` (override with `RUNLET_SHELL`). On Windows the default is `cmd.exe /C`.
- Working directory: `run.cwd` if absolute, `RUNLET_WORKSPACE/run.cwd` if relative, otherwise `RUNLET_WORKSPACE` directly.
- Environment: the runner inherits its own environment and appends the per-run `env` map on top.
- Timeout: `run.timeoutSeconds` if positive, otherwise `RUNLET_DEFAULT_TIMEOUT_SECONDS`. When the deadline is hit the runner sends `SIGTERM` to the entire process group, then reports status `timed_out`.
- Cancellation: pressing Ctrl-C or sending `SIGTERM` to the runner cancels the run (status `canceled`) and exits the polling loop.
- Output: stdout/stderr are scanned line-by-line. Each line becomes an event with `level=info` and `stream=stdout|stderr`. The runner also emits a `runner` stream event when the run starts and (on error) when something goes wrong.

## Operational notes

- The runner only supports `mode=shell`. A run with any other mode is finished as `failed` with an explanatory event.
- Concurrency is fixed at 1. Increase later versions only after Core adds per-runner queue isolation.
- The `setup` config file holds your bearer token in plaintext (mode `0600`). Treat it like an SSH private key.
- A runner identified by `(account_id, name)` is the same row across restarts; deleting it from `/<slug>/runners` and re-registering creates a brand new id.
