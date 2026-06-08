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
end
