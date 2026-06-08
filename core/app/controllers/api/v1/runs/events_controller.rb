class Api::V1::Runs::EventsController < Api::V1::BaseController
  before_action :set_runner_run

  def create
    @runner_run.record_event!(
      sequence: event_params[:sequence],
      level: event_params[:level],
      stream: event_params[:stream],
      message: event_params[:message],
      metadata: event_params[:metadata],
      occurred_at: event_params[:createdAt]
    )

    head :created
  end

  private
    def set_runner_run
      @runner_run = Current.account.runner_runs.find(params[:run_id])
    end

    def event_params
      params.permit(:sequence, :level, :stream, :message, :createdAt, metadata: {})
    end
end
