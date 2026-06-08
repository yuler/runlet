require "shellwords"

class RunnersController < ApplicationController
  def index
    @runners = Current.account.runners.includes(runner_runs: :events).order(last_heartbeat_at: :desc, created_at: :desc)
    @runner_api_url = root_url(script_name: "/#{Current.account.slug}").chomp("/")
    @runner_setup_token = Current.session&.signed_id || Current.identity.signed_id(purpose: :api_token)
    @runner_setup_command = [
      "runlet-runner",
      "setup",
      @runner_setup_token,
      "--api-url",
      @runner_api_url
    ].shelljoin
    @recent_runs = Current.account.runner_runs.includes(:runner, :events).recent.limit(10)
  end

  def destroy
    Current.account.runners.find(params[:id]).destroy
    redirect_to runners_path, notice: t(".removed")
  end
end
