# Development

Runlet is a pnpm monorepo with a Rails backend and a Go runner.

## Layout

```text
.
├── apps/runner/         # Go execution worker
├── core/                # Rails 8 backend + web UI
├── packages/
│   ├── brand/           # Logos, brand materials, static previews
│   ├── types/           # Shared TypeScript payload types
│   └── ui/              # Shared design CSS variables
├── docs/                # This documentation
├── scripts/             # Repo-wide automation
├── package.json         # Root pnpm scripts
└── pnpm-workspace.yaml
```

## Prerequisites

- Node.js (pnpm 9+)
- Ruby (see `core/.ruby-version`)
- Go 1.22+
- SQLite is bundled with macOS; on Linux install `libsqlite3-dev`.

`.agents/setup/SKILL.md` documents the in-house setup helper used by the agent fleet.

## Common commands

```bash
pnpm core:setup          # bin/rails db:prepare + brakeman/rubocop deps
pnpm core:dev            # bin/rails server -p 3000 -b "::"
pnpm core:test           # bin/rails test
pnpm core:commands       # list HTTP routes

pnpm runner:dev          # go run ./cmd/runlet-runner
pnpm runner:build        # go build -o ./bin/runlet-runner
pnpm runner:test         # go test ./...

pnpm verify:runner-flow  # scripts/verify-runner-flow.sh end-to-end smoke
pnpm types:check         # tsc on packages/types
pnpm build               # build all workspace projects
```

## Conventions

- Use **RTK** (`rtk <cmd>`) when running shell commands inside the agent fleet — see [`AGENTS.md`](../AGENTS.md). Outside the fleet, plain `bin/rails`, `go`, and `pnpm` work the same.
- Prefer `resource` / `resources` routes over standalone `get` / `post` actions (`AGENTS.md`).
- Account-scoped controllers rely on the `AccountSlug::Extractor` middleware: routes that should not be account-scoped must opt out with `script_name: nil`.
- UUIDv7 primary keys everywhere via `core/lib/rails_ext/active_record_uuid_type.rb`.
- The Rails app uses SQLite + Solid Queue / Cache / Cable for the MVP. No Redis or Postgres dependency.

## Adding apps

New client apps go under `apps/<name>` and expose commands through root scripts:

```json
"desktop:dev": "pnpm --filter @runlet/desktop dev"
```

Shared domain payload types belong in `packages/types`. UI-only helpers stay in the app until a second app needs them.

## Editor / hooks

The repo uses `.npmrc` and `.nvmrc` to pin pnpm and Node versions. Rubocop and Brakeman are wired through `core/bin/rubocop` and `core/bin/brakeman`.
