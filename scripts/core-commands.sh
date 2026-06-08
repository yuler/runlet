#!/usr/bin/env bash
set -euo pipefail

cat <<'COMMANDS'
Runlet core commands:

  pnpm core:setup        Install and prepare the Rails app
  pnpm core:dev          Start the Rails development server
  pnpm core:test         Run Rails tests
  pnpm core:lint         Run RuboCop
  pnpm core:rails -- ... Run Rails commands

Core API:

  DELETE /api/v1/session
  GET    /api/v1/session/heartbeat
  GET    /api/v1/stats
  POST   /api/v1/device/authorization
  POST   /api/v1/device/token
  POST   /api/v1/runners
  PATCH  /api/v1/runners/:id
  POST   /api/v1/runners/:runner_id/claims
COMMANDS
