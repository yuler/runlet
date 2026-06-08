class Api::V1::Runs::FinishesController < Api::V1::BaseController
  before_action :set_runner_run

  def create
    @runner_run.finish!(
      status: finish_params[:status],
      exit_code: finish_params[:exitCode],
      finished_at: finish_params[:finishedAt],
      message: finish_params[:message]
    )

    head :no_content
  end

  private
    def set_runner_run
      @runner_run = Current.account.runner_runs.find(params[:id])
    end

    def finish_params
      params.permit(:status, :exitCode, :finishedAt, :message)
    end
end
