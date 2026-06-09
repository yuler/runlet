require "test_helper"

class Api::V1::RunnerRunsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = Identity.create!(email: "runner-runs-api-#{SecureRandom.hex}@example.com")
    @account = Account.create!(name: "Runner Runs API", personal: true)
    @account.users.create!(identity: @identity, name: "Runner Owner", role: "owner")
    @session = @identity.sessions.create!(user_agent: "Runlet Runner/0.1.0", ip_address: "127.0.0.1")
    @headers = { Authorization: "Bearer #{@session.signed_id}" }
    @runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: {})
  end

  test "runner claims a queued shell run" do
    runner_run = @runner.runner_runs.create!(
      account: @account,
      identity: @identity,
      command: "pwd",
      cwd: "subdir",
      env: { "RAILS_ENV" => "test" },
      timeout_seconds: 30
    )

    post api_v1_runner_claims_url(@runner),
      params: { capacity: 1 },
      headers: @headers,
      as: :json

    assert_response :success
    json = JSON.parse(response.body)
    assert_equal runner_run.id, json.dig("run", "id")
    assert_equal "shell", json.dig("run", "mode")
    assert_equal "pwd", json.dig("run", "command")
    assert_equal "subdir", json.dig("run", "cwd")
    assert_equal({ "RAILS_ENV" => "test" }, json.dig("run", "env"))
    assert_equal 30, json.dig("run", "timeoutSeconds")
    assert_equal "running", runner_run.reload.status
  end

  test "runner records run events and finish state" do
    runner_run = @runner.runner_runs.create!(
      account: @account,
      identity: @identity,
      command: "printf hello"
    )
    runner_run.claim!(@runner)

    post api_v1_run_events_url(runner_run),
      params: {
        sequence: 1,
        level: "info",
        stream: "stdout",
        message: "hello",
        createdAt: Time.current.iso8601,
        metadata: { bytes: 5 }
      },
      headers: @headers,
      as: :json

    assert_response :created
    assert_equal "hello", runner_run.events.first.message

    post finish_api_v1_run_url(runner_run),
      params: {
        status: "succeeded",
        exitCode: 0,
        finishedAt: Time.current.iso8601,
        message: "run finished"
      },
      headers: @headers,
      as: :json

    assert_response :no_content
    assert_equal "succeeded", runner_run.reload.status
    assert_equal 0, runner_run.exit_code
  end

  test "claim returns null run when there is nothing queued" do
    post api_v1_runner_claims_url(@runner),
      params: { capacity: 1 },
      headers: @headers,
      as: :json

    assert_response :success
    assert_nil JSON.parse(response.body)["run"]
  end

  test "claim is scoped to the account" do
    other_identity = Identity.create!(email: "other-claim-#{SecureRandom.hex}@example.com")
    other_account = Account.create!(name: "Other Claim", personal: true)
    other_account.users.create!(identity: other_identity, name: "Owner", role: "owner")
    other_runner = Runner.register!(account: other_account, identity: other_identity, name: "local-runner", labels: {})

    post api_v1_runner_claims_url(other_runner),
      params: { capacity: 1 },
      headers: @headers,
      as: :json

    assert_response :not_found
  end

  test "event requires the run to belong to the authenticated account" do
    other_identity = Identity.create!(email: "other-event-#{SecureRandom.hex}@example.com")
    other_account = Account.create!(name: "Other Event", personal: true)
    other_account.users.create!(identity: other_identity, name: "Owner", role: "owner")
    other_runner = Runner.register!(account: other_account, identity: other_identity, name: "local-runner", labels: {})
    other_run = other_runner.runner_runs.create!(account: other_account, identity: other_identity, command: "true")

    post api_v1_run_events_url(other_run),
      params: { sequence: 1, level: "info", stream: "stdout", message: "x", createdAt: Time.current.iso8601 },
      headers: @headers,
      as: :json

    assert_response :not_found
  end

  test "event with duplicate sequence returns 422 and does not create a second event" do
    runner_run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")
    runner_run.claim!(@runner)
    runner_run.record_event!(sequence: 1, level: "info", stream: "stdout", message: "first", metadata: {}, occurred_at: Time.current)

    assert_no_difference "RunnerRunEvent.count" do
      post api_v1_run_events_url(runner_run),
        params: { sequence: 1, level: "info", stream: "stdout", message: "second", createdAt: Time.current.iso8601 },
        headers: @headers,
        as: :json
    end

    assert_response :unprocessable_entity
    assert_match(/sequence/i, JSON.parse(response.body)["error"])
  end

  test "finish a second time returns 422" do
    runner_run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")
    runner_run.claim!(@runner)
    runner_run.finish!(status: "succeeded", exit_code: 0, finished_at: nil, message: nil)

    post finish_api_v1_run_url(runner_run),
      params: { status: "failed", exitCode: 1, finishedAt: Time.current.iso8601, message: "second" },
      headers: @headers,
      as: :json

    assert_response :unprocessable_entity
    assert_equal "succeeded", runner_run.reload.status
  end

  test "finish updates an existing run" do
    runner_run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")
    runner_run.claim!(@runner)

    post finish_api_v1_run_url(runner_run),
      params: { status: "failed", exitCode: 1, finishedAt: Time.current.iso8601, message: "bad" },
      headers: @headers,
      as: :json

    assert_response :no_content
    runner_run.reload
    assert_equal "failed", runner_run.status
    assert_equal 1, runner_run.exit_code
    assert_equal "bad", runner_run.message
  end

  test "events require authentication" do
    runner_run = @runner.runner_runs.create!(account: @account, identity: @identity, command: "true")

    post api_v1_run_events_url(runner_run),
      params: { sequence: 1, level: "info", stream: "stdout", message: "x", createdAt: Time.current.iso8601 },
      as: :json

    assert_response :unauthorized
  end
end
