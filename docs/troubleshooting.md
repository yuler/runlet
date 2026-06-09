# Troubleshooting

Recovery steps and notable bug history for the MVP. Start here when something behaves unexpectedly during the [dispatch flow](dispatch-flow.md).

## Dropped output events (fixed)

**Symptom.** A shell run reports `status: succeeded` but the `runner_run_events` table is missing some of the lines the command wrote. The first event (`runner` / "run started") is always present, but stdout lines from the tail of the output are silently lost. Re-running the same command drops a different number of lines each time, which makes it look like a flaky test.

**Root cause.** `apps/runner/internal/executor/executor.go` used to wait for `cmd.Wait()` in a goroutine and join its result alongside the scan goroutines:

```go
// BUGGY pattern – do not reintroduce.
go func() { waitCh <- cmd.Wait() }()
wg.Wait()           // waited on scanners, but...
err := <-waitCh     // ...cmd.Wait() may have already returned
```

`cmd.Wait()` closes both `stdout` and `stderr` pipes the moment it returns. If the OS scheduled `cmd.Wait()` ahead of the scanner goroutines, the pipes were closed mid-read and `bufio.Scanner` ended its loop before draining the remaining buffered lines.

**Fix.** Drain the scanners first, then call `cmd.Wait()` synchronously:

```go
wg.Wait()                 // both stdout/stderr scanners have hit EOF
waitErr := cmd.Wait()     // safe to reap now – pipes already drained
```

This is the pattern the Go standard library docs recommend (see `os/exec` *"Wait will close the pipe after seeing the command exit"*).

**Regression test.** `apps/runner/internal/executor/executor_test.go::TestRunDoesNotDropOutputLines` runs a 200-line streaming command and asserts every line was emitted as an event. Keep it in place — it is the only thing that would have caught the original bug.

**Manual verification.** From the dispatch flow, queue a command that writes a known number of lines and assert the event count matches:

```bash
cd core
bin/rails runner '
run = RunnerRun.find("<run-id>")
expected = 1 + 200   # one "run started" + 200 stdout lines
abort "missing events: #{run.events.count} of #{expected}" unless run.events.count == expected
puts "OK: #{run.events.count} events"
'
```

If the count is short, the regression has come back. Bisect against the last known-good commit on `apps/runner/internal/executor/`.

## Core won't boot — stale PID file

**Symptom.** `pnpm core:dev` exits with `A server is already running. Check core/tmp/pids/server.pid` even though no Rails process exists.

**Fix.** Remove the stale file:

```bash
rm -f core/tmp/pids/server.pid
pnpm core:dev
```

This usually happens after a hard kill (Ctrl-\\, OOM, laptop sleep) — Rails leaves the pidfile behind.

## Runner registers but every claim returns `403` or `404`

**Symptom.** The runner logs `runner registered` once, then every poll prints something like `claim failed: status=404`.

**Likely causes.**

1. **Slug missing from `--api-url`.** The runner does *not* append the account slug for you. `http://localhost:3000` is wrong; `http://localhost:3000/<slug>` is right. Re-run `runlet-runner setup <token> --api-url http://localhost:3000/<slug>` to rewrite the config.
2. **Token belongs to a different account.** A `Session#signed_id` is bound to an identity, but the URL slug pins the request to an account. If the identity is not a user of that account, `Api::V1::BaseController` returns `403`. Confirm with:

   ```bash
   cd core
   bin/rails runner '
   identity = Identity.find_by!(email: "you@example.com")
   puts identity.users.map { |u| u.account.slug }
   '
   ```

3. **Runner row was deleted from `/<slug>/runners` while the binary still has the old `RUNLET_RUNNER_ID`.** Either re-run `setup` (which clears the saved id) or delete the `runner_id=` line from `runner.conf` by hand.

## `setup` succeeded but the runner says `token invalid`

The setup flow writes the token verbatim to disk; it does not validate it against Core. If you generated the token from a Rails console and copy-pasted it, double-check there is no whitespace or trailing newline:

```bash
grep -c '^token=' /tmp/runlet-runner-test/runner.conf   # should be 1
awk -F= '/^token=/ { print length($2) }' /tmp/runlet-runner-test/runner.conf
```

A valid `Session#signed_id` for this codebase is ~205 characters. Significantly shorter values are usually a truncated paste.

## `cmd.exe` errors on Windows

The MVP defaults to `cmd.exe /C` on Windows. Commands written for POSIX shells (pipes, `&&`, single-quote strings) will not behave the same. Either:

- override the shell at the runner level: `set RUNLET_SHELL=C:\Program Files\Git\bin\bash.exe`, or
- rewrite the queued command to be `cmd.exe`-compatible.

`apps/runner/internal/executor` only knows how to spawn one process and read its pipes; it does not parse the command itself.

## Heartbeat shows stale `last_heartbeat_at`

The runner heartbeats every `RUNLET_HEARTBEAT_INTERVAL_SECONDS` (default 15). If `/<slug>/runners` shows a heartbeat older than ~30s while the runner process is still up:

1. Check the runner logs for `heartbeat failed` lines — usually a network blip or Core restart.
2. Verify the runner can still reach Core: `curl -s -o /dev/null -w "%{http_code}\n" http://localhost:3000/up` from the same host.
3. If Core was restarted, the runner will recover on the next successful heartbeat; no action needed.

## Tearing down the verification environment

After running the [dispatch flow](dispatch-flow.md) against a real account, clean up so the next test starts from a known state:

```bash
cd core
bin/rails runner '
Account.find_by!(slug: "<slug>").runners.find_by!(name: "claude-test-runner")&.destroy
'
rm -rf /tmp/runlet-runner-test
```

Leaving the runner row in place is harmless, but its `last_heartbeat_at` will go stale and clutter the dashboard.
