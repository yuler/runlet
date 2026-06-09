# Dispatch flow

This is the end-to-end procedure we use to validate the MVP. Following it should reproduce a successful run from `queued` to `succeeded` with every output line captured. The same flow was used to find and fix the [dropped-output bug](troubleshooting.md#dropped-output-events-fixed).

## 1. Boot Core

```bash
pnpm core:setup            # only the first time, or after schema changes
pnpm core:dev              # or: cd core && bin/rails server -p 3000 -b "::"
```

Confirm Core is responding:

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:3000/up
# → 200
```

## 2. Mint a runner token

The runner authenticates with a `Session#signed_id`. Generate one for an existing identity that already owns the account you want to use:

```bash
cd core
bin/rails runner '
identity = Identity.find_by!(email: "you@example.com")
session = identity.sessions.create!(user_agent: "Runlet CLI", ip_address: "127.0.0.1")
puts session.signed_id
'
```

Note the account slug printed at `/<slug>/runners` — you will paste it into the runner's `--api-url`.

## 3. Configure and register the runner

```bash
runlet-runner setup <token> \
  --api-url 'http://localhost:3000/<slug>' \
  --workspace /tmp/runlet-runner-test \
  --name claude-test-runner \
  --config /tmp/runlet-runner-test/runner.conf
```

Run it once to perform the initial registration:

```bash
runlet-runner -config /tmp/runlet-runner-test/runner.conf -non-interactive -once
# → "runner registered" and the process exits if no run is queued
```

The runner is now visible on `/<slug>/runners` and has a stable `runner_id` reused on subsequent boots.

## 4. Queue a shell task

From the web UI: visit `/<slug>/runners`, fill in the "Queue shell run" form, submit.

Or from a Rails console:

```bash
cd core
bin/rails runner '
account = Account.find_by!(slug: "<slug>")
identity = Identity.find(account.users.where.not(identity_id: nil).first.identity_id)
runner = account.runners.find_by!(name: "claude-test-runner")
run = runner.runner_runs.create!(
  account: account,
  identity: identity,
  mode: "shell",
  command: "for i in 1 2 3 4 5; do echo line-$i; done"
)
puts "QUEUED: #{run.id}"
'
```

## 5. Execute and verify

```bash
runlet-runner -config /tmp/runlet-runner-test/runner.conf -non-interactive -once
```

Expected output (timestamps elided):

```
INFO msg="executing run" run_id=… cwd=/tmp/runlet-runner-test
INFO msg="run finished" run_id=… status=succeeded
```

Verify Core saw every event:

```bash
cd core
bin/rails runner '
run = RunnerRun.find("<run-id>")
puts "status: #{run.status}, events: #{run.events.count}"
run.events.order(:sequence).each { |e| puts "##{e.sequence} #{e.stream}: #{e.message.inspect}" }
'
```

You should see exactly `1 + N` events: one `runner` "run started" event followed by every output line, all with sequence numbers in order:

```
status: succeeded, events: 6
#1 runner: "run started"
#2 stdout: "line-1"
#3 stdout: "line-2"
#4 stdout: "line-3"
#5 stdout: "line-4"
#6 stdout: "line-5"
```

If any line is missing, see [troubleshooting](troubleshooting.md#dropped-output-events-fixed) — that bug was fixed in this version, but the regression test (`TestRunDoesNotDropOutputLines`) is the first place to look.

## 6. Tear down

```bash
# Stop the runner process (Ctrl-C if running long, otherwise it already exited via -once)
# Delete the runner row from the UI, or:
cd core
bin/rails runner '
Account.find_by!(slug: "<slug>").runners.find_by!(name: "claude-test-runner").destroy
'
rm -f /tmp/runlet-runner-test/runner.conf
```
