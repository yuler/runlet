# Runlet Core

`core/` is the Rails backend for the Runlet ecosystem.

It owns the durable model for runnable work:

- runlets: reusable task definitions
- runs: executions of a runlet
- schedules: recurring triggers
- events: observable state changes and logs
- tokens: app and CLI authentication

## Rails Boundary

The core app should stay responsible for persistence, scheduling, execution state, API authorization, and event delivery. Client-facing products in `apps/**` should call the HTTP API or subscribe to events; they should not duplicate domain rules.

## API Shape

Initial API namespace:

```text
/api/v1/runlets
/api/v1/runlets/:id/runs
/api/v1/runs/:id
/api/v1/runs/:id/events
/api/v1/schedules
```

Shared TypeScript types live in `packages/types` so apps can compile against the backend payload shape.

## Rails Defaults

When the Rails app is generated, prefer:

- Rails 8
- SQLite for local development
- Solid Queue for job execution
- Solid Cache
- Solid Cable for run event streaming
- UUID primary keys
