# Development

Runlet follows a simple monorepo interface:

```text
.
├── apps/              # Client-facing apps
├── core/              # Rails backend
├── packages/          # Shared packages
│   ├── brand/         # Brand assets and previews
│   └── types/         # TypeScript API payload types
├── scripts/           # Repo-wide automation
├── package.json       # Workspace scripts
└── pnpm-workspace.yaml
```

## Common Commands

```bash
pnpm core:setup
pnpm core:dev
pnpm core:test
pnpm types:check
```

## Adding Apps

Add deployable products under `apps/<name>` and expose commands through root scripts using pnpm filters:

```json
"desktop:dev": "pnpm --filter @runlet/desktop dev"
```

Shared domain payload types should go in `packages/types`; UI-only helpers should stay inside the app until a second app needs them.
