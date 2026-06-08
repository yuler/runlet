class Runners::RunsController < ApplicationController
  def create
    runner = Current.account.runners.find(params[:runner_id])
    runner.runner_runs.create!(
      account: Current.account,
      identity: Current.identity,
      mode: "shell",
      command: runner_run_params[:command],
      cwd: runner_run_params[:cwd].presence,
      timeout_seconds: runner_run_params[:timeout_seconds].presence
    )

    redirect_to runners_path, notice: t(".created")
  end

  private
    def runner_run_params
      params.require(:runner_run).permit(:command, :cwd, :timeout_seconds)
    end
end
