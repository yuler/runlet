# Runlet Runner

`apps/runner` is the first Go implementation of the Runlet execution worker.

The runner is intentionally small:

- register or reuse a local runner identity
- send heartbeats to Runlet Core
- claim one eligible run at a time
- execute the run command in a configured workspace
- stream stdout/stderr as run events
- finish the run with an exit code and status

## Development

```bash
go test ./...
go build -o ./bin/runlet-runner ./cmd/runlet-runner
go run ./cmd/runlet-runner -once
```

## Configuration

The default path is:

```bash
runlet-runner setup TOKEN --api-url http://localhost:3000/account-slug
runlet-runner
```

`setup` writes a small key/value config file to the OS user config directory. It does not use JSON config.

The runner also starts in inquire mode when no setup file exists and stdin is a terminal. It includes local test defaults, so pressing return through every prompt uses:

- API URL: `http://localhost:3000`
- token: `dev-token`
- name: `local-runner`
- workspace: the current directory
- labels: `kind=desktop,project=runlet`

Runtime config is merged as defaults, saved setup config, environment variables, then flags. Environment variables can override setup config:

```bash
RUNLET_API_URL=http://localhost:3000 \
RUNLET_TOKEN=dev-token \
RUNLET_WORKSPACE=/Users/yule/Projects/runlet \
RUNLET_LABELS=kind=desktop,project=runlet \
go run ./cmd/runlet-runner -once
```

Flags can override environment defaults:

```bash
go run ./cmd/runlet-runner \
  -api-url http://localhost:3000 \
  -token dev-token \
  -workspace /Users/yule/Projects/runlet \
  -label kind=desktop \
  -label project=runlet \
  -once
```

Use `-non-interactive` for scripts and CI. In that mode, required values must come from flags or environment variables.

Supported environment variables:

- `RUNLET_API_URL`
- `RUNLET_TOKEN`
- `RUNLET_RUNNER_ID`
- `RUNLET_RUNNER_NAME`
- `RUNLET_WORKSPACE`
- `RUNLET_SHELL`
- `RUNLET_LABELS`
- `RUNLET_POLL_INTERVAL_SECONDS`
- `RUNLET_HEARTBEAT_INTERVAL_SECONDS`
- `RUNLET_DEFAULT_TIMEOUT_SECONDS`

`RUNLET_SHELL` may be left empty to use the platform default (`/bin/sh` on Unix, `cmd.exe` on Windows).
