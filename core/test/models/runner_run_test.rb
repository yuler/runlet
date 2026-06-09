require "test_helper"

class RunnerRunTest < ActiveSupport::TestCase
  setup do
    @identity = Identity.create!(email: "runner-run-#{SecureRandom.hex}@example.com")
    @account = Account.create!(name: "Runner Run", personal: true)
    @account.users.create!(identity: @identity, name: "Runner Owner", role: "owner")
    @runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: {})
  end

  test "claim moves a queued shell run to running" do
    run = @runner.runner_runs.create!(
      account: @account,
      identity: @identity,
      command: "pwd"
    )

    claimed = run.claim!(@runner)

    assert_equal run, claimed
    assert_equal "running", run.status
    assert run.claimed_at.present?
    assert run.started_at.present?
  end

  test "claim ignores a run for another runner" do
    other = Runner.register!(account: @account, identity: @identity, name: "other-runner", labels: {})
    run = @runner.runner_runs.create!(
      account: @account,
      identity: @identity,
      command: "pwd"
    )

    assert_nil run.claim!(other)
    assert_equal "queued", run.reload.status
  end

  test "records runner output events" do
    run = @runner.runner_runs.create!(
      account: @account,
      identity: @identity,
      command: "printf hello"
    )

    run.record_event!(
      sequence: 1,
      level: "info",
      stream: "stdout",
      message: "hello",
      metadata: { bytes: 5 },
      occurred_at: Time.current
    )

    assert_equal "hello", run.events.first.message
    assert_equal({ "bytes" => "5" }, run.events.first.metadata)
  end

  test "claim is idempotent and does not re-running an already running run" do
    run = @runner.runner_runs.create!(
      account: @account,
      identity: @identity,
      command: "pwd"
    )

    assert_equal run, run.claim!(@runner)
    first_started = run.started_at

    assert_nil run.reload.claim!(@runner)
    assert_equal first_started.to_i, run.reload.started_at.to_i
  end

  test "record_event defaults level to info when blank" do
    run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")

    event = run.record_event!(sequence: 1, level: "", stream: nil, message: "hi", metadata: {}, occurred_at: nil)

    assert_equal "info", event.level
    assert_nil event.stream
    assert event.occurred_at.present?
  end

  test "finish writes status, exit code, and message" do
    run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")
    run.claim!(@runner)

    run.finish!(status: "succeeded", exit_code: 0, finished_at: nil, message: " all good ")

    assert_equal "succeeded", run.status
    assert_equal 0, run.exit_code
    assert_equal " all good ", run.message
    assert run.finished_at.present?
  end

  test "finish strips blank message to nil" do
    run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")
    run.claim!(@runner)

    run.finish!(status: "failed", exit_code: 1, finished_at: Time.current, message: "")

    assert_nil run.message
  end

  test "finish rejects non-terminal status" do
    run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")
    run.claim!(@runner)

    assert_raises(ActiveRecord::RecordInvalid) do
      run.finish!(status: "running", exit_code: nil, finished_at: nil, message: nil)
    end
  end

  test "finish refuses to overwrite an already-terminal run" do
    run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")
    run.claim!(@runner)
    run.finish!(status: "succeeded", exit_code: 0, finished_at: nil, message: nil)

    assert_raises(ActiveRecord::RecordInvalid) do
      run.finish!(status: "failed", exit_code: 1, finished_at: nil, message: "second attempt")
    end

    assert_equal "succeeded", run.reload.status
    assert_equal 0, run.exit_code
  end

  test "claimable scope returns queued runs ordered by created_at asc" do
    older = @runner.runner_runs.create!(account: @account, identity: @identity, command: "echo 1", created_at: 2.minutes.ago)
    newer = @runner.runner_runs.create!(account: @account, identity: @identity, command: "echo 2")
    running = @runner.runner_runs.create!(account: @account, identity: @identity, command: "echo 3")
    running.update!(status: "running")

    claimable = @runner.runner_runs.claimable.to_a

    assert_equal [ older, newer ], claimable
  end

  test "rejects invalid mode" do
    run = @runner.runner_runs.build(account: @account, identity: @identity, command: "true", mode: "docker")

    assert_not run.valid?
    assert_includes run.errors[:mode], "is not included in the list"
  end

  test "rejects blank command" do
    run = @runner.runner_runs.build(account: @account, identity: @identity, command: "")

    assert_not run.valid?
    assert_includes run.errors[:command], "can't be blank"
  end

  test "rejects invalid status" do
    run = @runner.runner_runs.build(account: @account, identity: @identity, command: "true", status: "weird")

    assert_not run.valid?
    assert_includes run.errors[:status], "is not included in the list"
  end

  test "rejects nested env values" do
    run = @runner.runner_runs.build(account: @account, identity: @identity, command: "true", env: { "nested" => { "x" => 1 } })

    assert_not run.valid?
    assert_includes run.errors[:env], "must be a flat object with string values"
  end

  test "destroying a run cascades its events" do
    run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")
    run.record_event!(sequence: 1, level: "info", stream: "stdout", message: "ok", metadata: {}, occurred_at: Time.current)

    assert_difference "RunnerRunEvent.count", -1 do
      run.destroy
    end
  end
end
