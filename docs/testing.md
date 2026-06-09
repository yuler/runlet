# Testing

The MVP has two test suites — Rails (Minitest) for Core and Go for the runner — both wired up to run in parallel.

## Run everything

```bash
pnpm core:test                     # Rails
cd apps/runner && go test ./...    # Runner
```

Both suites are deterministic, hermetic, and require no external services. The current totals:

| Suite           | Tests | Assertions / Subtests |
| --------------- | ----- | --------------------- |
| `core/`         | 92    | 297                   |
| `apps/runner/`  | 43    | n/a (table-driven)    |

## Rails (`core/`)

Layout:

```
core/test/
├── controllers/
│   ├── api/v1/{runners,runner_runs,stats,sessions}_controller_test.rb
│   ├── runners_controller_test.rb       # browser flow + queue form
│   ├── dashboards_controller_test.rb
│   └── …
├── integration/
│   └── device_authorization_flow_test.rb
├── models/
│   ├── runner_test.rb                   # registration, heartbeat, validations
│   ├── runner_run_test.rb               # claim, finish, scopes, validations
│   ├── runner_run_event_test.rb         # sequence uniqueness, level/stream constraints
│   ├── session_test.rb
│   ├── signup_test.rb
│   └── device/authorization_test.rb
└── test_helper.rb                       # FixturesUuidHelper + parallel workers
```

Highlights worth knowing about:

- Tests run with `parallelize(workers: :number_of_processors)`. Fixtures get deterministic UUIDv7 values via `FixturesUuidHelper` so `.first` / `.last` stay stable.
- Setup helpers in controller tests create their own identity / account / session so they don't depend on shared fixtures.
- Account-scoped routes are reached with `script_name: "/#{@account.slug}"` — tests that hit `/<slug>/...` need this.

### Adding a Rails test

1. Add a model test in `core/test/models/<model>_test.rb` covering every public method plus invalid-state branches.
2. Add a controller test in `core/test/controllers/...` that exercises the route through `ActionDispatch::IntegrationTest`. Always include an unauthorized / wrong-account scenario.
3. Run `bin/rails test test/path/to/specific_test.rb` while iterating; run the full suite before committing.

## Go (`apps/runner/`)

Layout:

```
apps/runner/
├── cmd/runlet-runner/main_test.go        # setup CLI behavior
└── internal/
    ├── api/client_test.go                # httptest-backed Client
    ├── config/config_test.go             # seeds, env, prompts, validation
    ├── config/store_test.go              # JSON settings read/write
    ├── daemon/daemon_test.go             # background runner start
    ├── executor/executor_test.go         # shell execution, timeouts, env, no-drop regression
    └── runner/runner_test.go             # registration, claim+execute happy path
```

Notable patterns:

- `httptest.NewServer` for API client tests; never spin up real Core.
- `t.TempDir()` for any workspace or config file.
- `t.Setenv` (with empty values) is used to scrub `RUNLET_*` from the environment so the test is hermetic.
- `runner/runner_test.go` uses a `fakeAPI` that satisfies the `API` interface — replace its `runs` slice to drive scenarios.
- The executor suite includes `TestRunDoesNotDropOutputLines` — a 200-line streaming command — which is the regression guard for the dropped-events bug fixed in this version (see [troubleshooting](troubleshooting.md#dropped-output-events-fixed)).

### Adding a Go test

1. Place new tests next to the file they cover (`<file>_test.go`).
2. Use the existing `fakeAPI` for service-level tests; only reach for `httptest` when you're testing the API client itself.
3. `go test ./...` from `apps/runner/` runs the full suite in under 2 seconds.

## End-to-end verification

The script `scripts/verify-runner-flow.sh` (run via `pnpm verify:runner-flow`) provisions a temporary token, registers a runner, queues a run, and asserts the events were captured. The same procedure performed by hand is documented in [`dispatch-flow.md`](dispatch-flow.md).
