require "test_helper"

class RunnerTest < ActiveSupport::TestCase
  setup do
    @identity = Identity.create!(email: "runner-model-#{SecureRandom.hex}@example.com")
    @account = Account.create!(name: "Runner Model", personal: true)
    @account.users.create!(identity: @identity, name: "Runner Owner", role: "owner")
  end

  test "register creates a runner scoped to an account and identity" do
    runner = Runner.register!(
      account: @account,
      identity: @identity,
      name: " local-runner ",
      labels: { arch: :arm64, os: "darwin" }
    )

    assert_equal @account, runner.account
    assert_equal @identity, runner.identity
    assert_equal "local-runner", runner.name
    assert_equal({ "arch" => "arm64", "os" => "darwin" }, runner.labels)
    assert_equal "idle", runner.status
    assert runner.last_heartbeat_at.present?
  end

  test "register reuses runner names within the same account" do
    runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: { os: "darwin" })

    assert_no_difference "Runner.count" do
      reused = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: { os: "linux" })
      assert_equal runner.id, reused.id
      assert_equal({ "os" => "linux" }, reused.labels)
    end
  end

  test "heartbeat records current status" do
    runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: {})

    assert_changes -> { runner.reload.last_heartbeat_at } do
      runner.heartbeat!(status: "running", current_run_id: "run_123", labels: { queue: "default" })
    end

    assert_equal "running", runner.status
    assert_equal "run_123", runner.current_run_id
    assert_equal({ "queue" => "default" }, runner.labels)
  end

  test "heartbeat falls back to idle when status is blank" do
    runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: {})

    runner.heartbeat!(status: "", current_run_id: nil, labels: {})

    assert_equal "idle", runner.status
    assert_nil runner.current_run_id
  end

  test "heartbeat ignores blank current_run_id" do
    runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: {})

    runner.heartbeat!(status: "running", current_run_id: "", labels: {})

    assert_nil runner.current_run_id
  end

  test "register coerces non-string label values into strings" do
    runner = Runner.register!(
      account: @account,
      identity: @identity,
      name: "local-runner",
      labels: { capacity: 4, busy: false }
    )

    assert_equal({ "capacity" => "4", "busy" => "false" }, runner.labels)
  end

  test "rejects invalid status" do
    runner = Runner.new(account: @account, identity: @identity, name: "x", status: "bogus", labels: {})

    assert_not runner.valid?
    assert_includes runner.errors[:status], "is not included in the list"
  end

  test "rejects blank name" do
    runner = Runner.new(account: @account, identity: @identity, name: "   ", status: "idle", labels: {})

    assert_not runner.valid?
    assert_includes runner.errors[:name], "can't be blank"
  end

  test "rejects nested label values" do
    runner = Runner.new(
      account: @account,
      identity: @identity,
      name: "local-runner",
      status: "idle",
      labels: { meta: { nested: "value" } }
    )

    assert_not runner.valid?
    assert_includes runner.errors[:labels], "must be a flat object with string values"
  end

  test "enforces unique runner names within an account" do
    Runner.create!(account: @account, identity: @identity, name: "local-runner", labels: {})

    duplicate = Runner.new(account: @account, identity: @identity, name: "local-runner", labels: {})

    assert_raises(ActiveRecord::RecordNotUnique) { duplicate.save(validate: false) }
  end

  test "for_account scope filters by account" do
    other_identity = Identity.create!(email: "scope-#{SecureRandom.hex}@example.com")
    other_account = Account.create!(name: "Scoped", personal: true)
    other_account.users.create!(identity: other_identity, name: "Owner", role: "owner")
    mine = Runner.register!(account: @account, identity: @identity, name: "mine", labels: {})
    Runner.register!(account: other_account, identity: other_identity, name: "theirs", labels: {})

    assert_equal [ mine ], Runner.for_account(@account).to_a
  end
end
