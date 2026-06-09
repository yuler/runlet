# Runlet Documentation

Runlet is a lightweight platform for defining small pieces of work and dispatching them to trusted runner machines. This directory is the single source of truth for how the system fits together.

The current release is the **MVP**: Core stores accounts, runners, and shell run requests; the Go runner registers, heartbeats, claims one shell run at a time, executes it in a configured workspace, and streams events back.

## Contents

- [`architecture.md`](architecture.md) — system overview, how Core, the runner, and shared packages interact.
- [`api.md`](api.md) — HTTP API reference used by the runner and external clients.
- [`runner.md`](runner.md) — installing, configuring, and operating the Go runner.
- [`dispatch-flow.md`](dispatch-flow.md) — end-to-end walkthrough of queueing a run and verifying its execution.
- [`development.md`](development.md) — local environment setup, common commands, repository layout.
- [`testing.md`](testing.md) — how the test suites are organized and how to run them.
- [`troubleshooting.md`](troubleshooting.md) — known issues, recovery steps, and recently-fixed bugs.

## Quick start

```bash
# 1. Boot Core
pnpm core:setup
pnpm core:dev

# 2. Generate a runner token from a Rails console
cd core
bin/rails runner 'session = Identity.find_by!(email: "you@example.com").sessions.create!(user_agent: "Runlet CLI", ip_address: "127.0.0.1"); puts session.signed_id'

# 3. Configure and start the runner
runlet-runner setup <token> --api-url http://localhost:3000/<account-slug>
runlet-runner -once   # or omit -once to keep polling
```

See [`dispatch-flow.md`](dispatch-flow.md) for the full hands-on walkthrough used to validate the MVP.
