---
name: setup
description: Use this when preparing, updating, or verifying the local Runlet development environment, including pnpm workspace dependencies, the Rails core app, shared TypeScript types, and repo-level health checks.
---

# Runlet Setup

Use this skill when the user asks to set up this repo, refresh dependencies, start local development, or verify the workspace after environment changes.

## Project Shape

Runlet is a pnpm monorepo with:

- `core/`: Rails backend
- `packages/types/`: shared TypeScript payload types
- `packages/brand/`: brand assets and static previews
- `apps/`: future client-facing apps

Respect the repo instruction to prefix shell commands with `rtk`.

## Environment

- Node: `24` from `.nvmrc`
- Package manager: `pnpm@10.33.0` from root `package.json`
- Ruby: `4.0.3` from `core/.ruby-version`
- Rails app database: SQLite, prepared by `core/bin/setup`

## Setup Workflow

From the repository root:

```bash
rtk pnpm install
rtk pnpm core:setup -- --skip-server
rtk pnpm types:check
```

Notes:

- `pnpm core:setup` runs `core/bin/setup`.
- `core/bin/setup` installs Ruby gems, prepares the Rails database, clears logs/tempfiles, and starts the dev server unless `--skip-server` is passed.
- Use `--skip-server` for non-interactive setup, CI-like checks, or when the user only asked to prepare the environment.

## Development Commands

```bash
rtk pnpm core:dev
rtk pnpm core:test
rtk pnpm core:lint
rtk pnpm core:lint:fix
rtk pnpm types:check
rtk pnpm core:commands
```

Use `pnpm core:dev` to start the Rails server after setup.

## Verification

For a quick health check after setup or dependency changes:

```bash
rtk pnpm types:check
rtk pnpm core:test
```

Run `rtk pnpm core:lint` when Ruby style or Rails code changed.

## Troubleshooting

- If pnpm fails because of engine strictness, switch to Node 24 before installing.
- If Ruby gems fail to install, confirm Ruby 4.0.3 is active inside `core/`.
- If the Rails database is stale or missing, rerun `rtk pnpm core:setup -- --skip-server`.
