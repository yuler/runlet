# HTTP API

All routes live under `/<account-slug>/api/v1/`. Every request must include:

- `Authorization: Bearer <token>` — either a `Session#signed_id` or an identity access token.
- `Content-Type: application/json` and `Accept: application/json` for write operations.

The base controller renders `404` for `ActiveRecord::RecordNotFound`, `422` for `ActiveRecord::RecordInvalid` (validation failures and state-machine guards), `401` for missing authentication, `403` when the authenticated identity is not a user of the URL's account, and `426 Upgrade Required` when an outdated `Runlet Desktop/x.y.z` user agent is detected (current minimum `1.6.0`).

> Examples below assume the account slug `acme` and a runner id `rnr_xyz`.

## Runners

### Register / reuse a runner

```
POST /acme/api/v1/runners
```

```json
{ "name": "local-runner", "labels": { "os": "darwin", "arch": "arm64" } }
```

- 201 `{ "runnerId": "<uuid>" }` — A runner with the given name is created if missing, otherwise the existing runner is returned and its labels are overwritten. Names are unique per account.
- 422 `{ "error": "<message>" }` — Validation error (e.g. blank name).

### Heartbeat

```
PATCH /acme/api/v1/runners/rnr_xyz
```

```json
{ "status": "running", "currentRunId": "run_abc", "labels": { "queue": "default" } }
```

- 204 — heartbeat accepted. Empty `status` falls back to `"idle"`.
- 404 — runner does not belong to the authenticated account.

### Claim the next queued run

```
POST /acme/api/v1/runners/rnr_xyz/claims
```

```json
{ "capacity": 1, "labels": { "os": "darwin" } }
```

- 200 `{ "run": null }` — nothing to claim.
- 200 `{ "run": { … } }` — the oldest queued run for this runner was moved to `running`. Payload:

  ```json
  {
    "id": "run_abc",
    "runletId": "run_abc",
    "mode": "shell",
    "command": "pwd",
    "cwd": "subdir-or-absolute",
    "env": { "RAILS_ENV": "test" },
    "timeoutSeconds": 30
  }
  ```

  `cwd` is optional. When it is relative, the runner joins it with its configured `RUNLET_WORKSPACE`.

## Runs

### Stream an event

```
POST /acme/api/v1/runs/run_abc/events
```

```json
{
  "sequence": 7,
  "level": "info",
  "stream": "stdout",
  "message": "hello from runlet",
  "createdAt": "2026-06-09T15:30:00Z",
  "metadata": { "bytes": "12" }
}
```

- 201 — event stored.
- 404 — run does not belong to the authenticated account.
- 422 — validation failure: `sequence` must be unique within the run, `level` must be one of `debug|info|warn|error`, `stream` must be empty or `runner|stdout|stderr`.

### Finish a run

```
POST /acme/api/v1/runs/run_abc/finish
```

```json
{
  "status": "succeeded",
  "exitCode": 0,
  "finishedAt": "2026-06-09T15:30:01Z",
  "message": "run finished"
}
```

- 204 — terminal state recorded. `status` must be one of `succeeded|failed|timed_out|canceled` (non-terminal values return 422).
- 422 — the run is already in a terminal state, or `status` is missing/non-terminal. The previous finish is left untouched.

## Sessions and devices (used by the web/desktop flows)

- `DELETE /acme/api/v1/session` — sign out the current session.
- `GET /acme/api/v1/session/heartbeat` — keep-alive ping.
- `POST /acme/api/v1/device/authorization` — start a device-code authorization request.
- `POST /acme/api/v1/device/token` — exchange a device code for a session token after the user approves.

## Stats

- `GET /acme/api/v1/stats` — high-level account counters used by the dashboard.

## Conventions

- Timestamps are ISO 8601 in UTC.
- IDs are UUIDv7 strings encoded in base36 (25 chars), e.g. `03ga7ac6w4enw8k7og8bu41ya`.
- Empty optional fields are omitted from request bodies (the Go client follows this rule via `omitempty`).
