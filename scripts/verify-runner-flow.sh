#!/usr/bin/env bash
set -euo pipefail

(
  cd core
  bin/rails test \
    test/models/runner_test.rb \
    test/models/runner_run_test.rb \
    test/controllers/runners_controller_test.rb \
    test/controllers/api/v1/runners_controller_test.rb \
    test/controllers/api/v1/runner_runs_controller_test.rb
)

pnpm runner:test
pnpm types:check
