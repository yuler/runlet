# Runlet

**Let it run.**

Runlet is a lightweight runner platform for defining small pieces of work, sending them to trusted machines, and watching each run move from claim to completion.

Runlet keeps the core loop intentionally simple: Core owns accounts, run definitions, schedules, runner registration, and execution state; runners connect back to Core, heartbeat, claim work, execute commands in a configured workspace, and stream events.

![Runlet logo](packages/brand/logos/runlet-logo-clean-01-primary-lockup.png)

## Slogan

**Let it run.**

The slogan is short enough to work as both a product promise and an operator command: define the work, hand it to Runlet, and let the runner carry it reliably.

Supporting lines:

- Tiny jobs. Reliable runs.
- Make work runnable.
- Run the work, not the worry.

## Summary

Runlet is built around three pieces:

- **Core**: a Rails app that owns identity, accounts, runlets, runs, schedules, runners, and API state.
- **Runner**: a small Go worker that registers itself with Core, sends heartbeats, claims one run at a time, executes it, and reports events.
- **Packages**: shared TypeScript types, UI design tokens, and brand assets used across future apps.

The product direction is developer-first infrastructure: precise, observable, calm, and small enough to understand locally.

## Monorepo Structure

```text
.
├── apps/
│   └── runner/        # Go execution worker
├── core/              # Rails backend and web UI
├── packages/
│   ├── brand/         # Logos, brand materials, and static previews
│   ├── types/         # Shared TypeScript payload types
│   └── ui/            # Shared design CSS variables
├── scripts/           # Repo-wide automation
├── package.json       # Root pnpm scripts
└── pnpm-workspace.yaml
```

Common commands:

```bash
pnpm core:setup
pnpm core:dev
pnpm core:test
pnpm runner:dev
pnpm runner:build
pnpm runner:test
pnpm verify:runner-flow
pnpm types:check
pnpm build
```

Core API routes can be listed with:

```bash
pnpm core:commands
```

## Register a Runner

Start Core first:

```bash
pnpm core:setup
pnpm core:dev
```

Create or sign in to an account in the local Rails app. For local development, generate a temporary runner token from the Rails console:

```bash
cd core
bin/rails runner 'identity = Identity.find_by!(email: "you@example.com"); puts identity.signed_id(purpose: :api_token)'
```

Then start the runner from the repo root. The runner will register itself on first boot, then reuse the same runner name on later boots:

```bash
RUNLET_API_URL=http://localhost:3000 \
RUNLET_TOKEN=<token> \
RUNLET_RUNNER_NAME=local-runner \
RUNLET_WORKSPACE=/Users/yule/Projects/runlet \
RUNLET_LABELS=kind=desktop,project=runlet \
pnpm runner:dev
```

Once it has started, the runner should appear on the account runners page at `/<account-slug>/runners`.
From that page, queue a shell command for the runner. The first version is intentionally shell-only: Core stores the command, the runner claims one queued command for its configured workspace, executes it through the local shell, streams stdout/stderr back to Core, and records the final exit code.

You can also pass the same values as flags:

```bash
cd apps/runner
go run ./cmd/runlet-runner \
  -api-url http://localhost:3000 \
  -token <token> \
  -name local-runner \
  -workspace /Users/yule/Projects/runlet \
  -label kind=desktop \
  -label project=runlet
```

Use `-once` to claim and execute at most one run, or `-non-interactive` when the runner is launched by scripts or CI. In non-interactive mode, required values must come from flags or environment variables.

Supported runner environment variables:

- `RUNLET_API_URL`
- `RUNLET_TOKEN`
- `RUNLET_RUNNER_ID`
- `RUNLET_RUNNER_NAME`
- `RUNLET_WORKSPACE`
- `RUNLET_SHELL`
- `RUNLET_LABELS`
- `RUNLET_CONCURRENCY`
- `RUNLET_POLL_INTERVAL_SECONDS`
- `RUNLET_HEARTBEAT_INTERVAL_SECONDS`
- `RUNLET_DEFAULT_TIMEOUT_SECONDS`

Current runner behavior:

- registers or reuses a runner by account and name
- sends periodic heartbeats to Core
- claims one queued shell run at a time
- executes commands in the configured workspace unless the run provides a relative override
- streams stdout and stderr as run events
- reports final exit code and status

## Brand Assets

Current logo assets live in [`packages/brand/logos`](packages/brand/logos/).

Preview the clean PNG logo set:

[`packages/brand/logos/preview.html`](packages/brand/logos/preview.html)

Generated brand material concepts live in [`packages/brand/materials`](packages/brand/materials/).

Preview all material concepts:

[`packages/brand/materials/preview.html`](packages/brand/materials/preview.html)

Detailed visual rules are documented in [`DESIGN.md`](DESIGN.md).

Preview the visual style guide:

[`packages/brand/design/preview.html`](packages/brand/design/preview.html)
