require "test_helper"

class Api::V1::RunnersControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = Identity.create!(email: "runner-api-#{SecureRandom.hex}@example.com")
    @account = Account.create!(name: "Runner API", personal: true)
    @account.users.create!(identity: @identity, name: "Runner Owner", role: "owner")
    @session = @identity.sessions.create!(user_agent: "Runlet Runner/0.1.0", ip_address: "127.0.0.1")
    @headers = { Authorization: "Bearer #{@session.signed_id}" }
  end

  test "register creates a runner" do
    assert_difference "Runner.count", 1 do
      post api_v1_runners_url,
        params: { name: "local-runner", labels: { os: "darwin", arch: "arm64" } },
        headers: @headers,
        as: :json
    end

    assert_response :created
    json = JSON.parse(response.body)
    runner = Runner.find(json["runnerId"])
    assert_equal @account, runner.account
    assert_equal @identity, runner.identity
    assert_equal({ "os" => "darwin", "arch" => "arm64" }, runner.labels)
  end

  test "register reuses an existing runner for the same account and name" do
    runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: { os: "darwin" })

    assert_no_difference "Runner.count" do
      post api_v1_runners_url,
        params: { name: "local-runner", labels: { os: "linux" } },
        headers: @headers,
        as: :json
    end

    assert_response :created
    assert_equal runner.id, JSON.parse(response.body)["runnerId"]
    assert_equal({ "os" => "linux" }, runner.reload.labels)
  end

  test "heartbeat updates runner state" do
    runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: {})

    assert_changes -> { runner.reload.last_heartbeat_at } do
      patch api_v1_runner_url(runner),
        params: { status: "running", currentRunId: "run_123", labels: { queue: "default" } },
        headers: @headers,
        as: :json
    end

    assert_response :no_content
    assert_equal "running", runner.status
    assert_equal "run_123", runner.current_run_id
    assert_equal({ "queue" => "default" }, runner.labels)
  end

  test "heartbeat is scoped to the authenticated account" do
    other_identity = Identity.create!(email: "other-runner-#{SecureRandom.hex}@example.com")
    other_account = Account.create!(name: "Other Runner API", personal: true)
    other_account.users.create!(identity: other_identity, name: "Other Owner", role: "owner")
    other_runner = Runner.register!(account: other_account, identity: other_identity, name: "local-runner", labels: {})

    patch api_v1_runner_url(other_runner),
      params: { status: "idle", labels: {} },
      headers: @headers,
      as: :json

    assert_response :not_found
  end

  test "claim returns no run until shell run exists" do
    runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: {})

    post api_v1_runner_claims_url(runner),
      params: { capacity: 1, labels: { os: "darwin" } },
      headers: @headers,
      as: :json

    assert_response :success
    assert_nil JSON.parse(response.body)["run"]
  end

  test "register requires authentication" do
    post api_v1_runners_url,
      params: { name: "local-runner", labels: {} },
      as: :json

    assert_response :unauthorized
  end
end
