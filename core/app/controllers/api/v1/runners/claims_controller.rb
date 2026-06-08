class Api::V1::Runners::ClaimsController < Api::V1::BaseController
  before_action :set_runner

  def create
    runner_run = @runner.runner_runs.claimable.first&.claim!(@runner)

    render json: { run: runner_run && run_payload(runner_run) }
  end

  private
    def set_runner
      @runner = Current.account.runners.find(params[:runner_id])
    end

    def run_payload(runner_run)
      {
        id: runner_run.id,
        runletId: runner_run.id,
        mode: runner_run.mode,
        command: runner_run.command,
        cwd: runner_run.cwd,
        env: runner_run.env,
        timeoutSeconds: runner_run.timeout_seconds.to_i
      }
    end
end
