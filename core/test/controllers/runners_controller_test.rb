require "test_helper"

class RunnersControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = Identity.create!(email: "runners-page-#{SecureRandom.hex}@example.com")
    @account = Account.create!(name: "Runners Page", personal: true)
    @account.users.create!(identity: @identity, account: @account, name: "Runner Owner", role: "owner")
    @session = @identity.sessions.create!(user_agent: "TestAgent", ip_address: "127.0.0.1")
  end

  test "should list account runners" do
    Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: { os: "darwin" })
    sign_in

    get runners_url(script_name: "/#{@account.slug}")

    assert_response :success
    assert_select "h1", "Runners"
    assert_select "h2", "Connect a runner"
    assert_select "code", text: /runlet-runner setup .+ --api-url http:\/\/www.example.com\/#{@account.slug}/
    assert_select "code", text: "runlet-runner"
    assert_select "div", text: /local-runner/
    assert_select "span", text: "os=darwin"
    assert_select "textarea[name='runner_run[command]']"
  end

  test "should queue shell run for runner" do
    runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: {})
    sign_in

    assert_difference "RunnerRun.count", 1 do
      post runner_runs_url(runner, script_name: "/#{@account.slug}"),
        params: { runner_run: { command: "pwd", cwd: "subdir", timeout_seconds: 30 } }
    end

    run = RunnerRun.last
    assert_equal @account, run.account
    assert_equal runner, run.runner
    assert_equal @identity, run.identity
    assert_equal "shell", run.mode
    assert_equal "queued", run.status
    assert_equal "pwd", run.command
    assert_redirected_to runners_path
    assert_equal "Shell run queued.", flash[:notice]
  end

  test "should show empty state" do
    sign_in

    get runners_url(script_name: "/#{@account.slug}")

    assert_response :success
    assert_select "div", text: I18n.t("runners.index.empty")
  end

  test "should remove runner" do
    runner = Runner.register!(account: @account, identity: @identity, name: "local-runner", labels: {})
    sign_in

    assert_difference "Runner.count", -1 do
      delete runner_url(runner, script_name: "/#{@account.slug}")
    end

    assert_redirected_to runners_path
    assert_equal "Runner removed.", flash[:notice]
  end

  private
    def sign_in
      post session_url, params: { email: @identity.email }
      magic_link = @identity.magic_links.last
      post session_magic_link_url, params: { code: magic_link.code }
    end
end
