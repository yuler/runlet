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
end
