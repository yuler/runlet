# Agents

Runlet is a pnpm monorepo for a Rails core app, shared TypeScript packages, brand assets, and future apps.

## Rules

- Use RTK when running shell commands. Run from the repository root and keep the required `rtk` prefix.
- Use `gh` operate on GitHub issues and pull requests.

## Local Skills

- Use `.agents/setup/SKILL.md` when preparing, updating, or verifying the local development environment.

## Project Layout

```text
.
├── apps/              # Client-facing apps
├── core/              # Rails backend
├── packages/
│   ├── brand/         # Brand assets and static previews
│   ├── types/         # Shared TypeScript payload types
│   └── ui/            # Shared design CSS variables
└── scripts/           # Repo-wide automation
```

## Core

### Account Slug

- `core/config/initializers/account_slug.rb` installs `AccountSlug::Extractor` middleware. It treats a non-reserved first path segment as an account slug, moves that segment into `SCRIPT_NAME`, and runs the request with `Current.account`.
- Routes that do not require an account context must explicitly declare `script_name: nil` so they bypass account slug mounting and do not get accidentally scoped to an account.

### Routes

- Prefer `resource` / `resources` routes over standalone `get` / `post` action routes.
