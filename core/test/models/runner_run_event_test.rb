require "test_helper"

class RunnerRunEventTest < ActiveSupport::TestCase
  setup do
    @identity = Identity.create!(email: "event-#{SecureRandom.hex}@example.com")
    @account = Account.create!(name: "Event Account", personal: true)
    @account.users.create!(identity: @identity, name: "Owner", role: "owner")
    @runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: {})
    @run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")
  end

  test "requires sequence, level, message, occurred_at" do
    event = RunnerRunEvent.new(runner_run: @run)

    assert_not event.valid?
    assert_includes event.errors[:sequence], "can't be blank"
    assert_includes event.errors[:message], "can't be blank"
    assert_includes event.errors[:occurred_at], "can't be blank"
  end

  test "rejects unknown level" do
    event = RunnerRunEvent.new(
      runner_run: @run,
      sequence: 1,
      level: "verbose",
      message: "x",
      occurred_at: Time.current
    )

    assert_not event.valid?
    assert_includes event.errors[:level], "is not included in the list"
  end

  test "rejects unknown stream" do
    event = RunnerRunEvent.new(
      runner_run: @run,
      sequence: 1,
      level: "info",
      stream: "unknown",
      message: "x",
      occurred_at: Time.current
    )

    assert_not event.valid?
    assert_includes event.errors[:stream], "is not included in the list"
  end

  test "allows blank stream" do
    event = RunnerRunEvent.new(
      runner_run: @run,
      sequence: 1,
      level: "info",
      stream: "",
      message: "x",
      occurred_at: Time.current
    )

    assert event.valid?, event.errors.full_messages.to_sentence
  end

  test "sequence must be unique within a run" do
    @run.record_event!(sequence: 1, level: "info", stream: "stdout", message: "first", metadata: {}, occurred_at: Time.current)

    duplicate = RunnerRunEvent.new(
      runner_run: @run,
      sequence: 1,
      level: "info",
      stream: "stdout",
      message: "second",
      occurred_at: Time.current
    )

    assert_not duplicate.valid?
    assert_includes duplicate.errors[:sequence], "has already been taken"
  end

  test "sequence can repeat across different runs" do
    @run.record_event!(sequence: 1, level: "info", stream: "stdout", message: "first", metadata: {}, occurred_at: Time.current)
    other_run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")

    assert_nothing_raised do
      other_run.record_event!(sequence: 1, level: "info", stream: "stdout", message: "first", metadata: {}, occurred_at: Time.current)
    end
  end
end
