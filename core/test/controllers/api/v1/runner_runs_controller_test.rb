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
end
