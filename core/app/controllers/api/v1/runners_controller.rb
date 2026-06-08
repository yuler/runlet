class Api::V1::RunnersController < Api::V1::BaseController
  before_action :set_runner, only: :update

  rescue_from ActiveRecord::RecordInvalid do |exception|
    render json: { error: exception.record.errors.full_messages.to_sentence }, status: :unprocessable_entity
  end

  def create
    runner = Runner.register!(
      account: Current.account,
      identity: Current.identity,
      name: runner_params[:name],
      labels: runner_params[:labels]
    )

    render json: { runnerId: runner.id }, status: :created
  end

  def update
    @runner.heartbeat!(
      status: heartbeat_params[:status],
      current_run_id: heartbeat_params[:currentRunId],
      labels: heartbeat_params[:labels]
    )

    head :no_content
  end

  private
    def set_runner
      @runner = Current.account.runners.find(params[:id])
    end

    def runner_params
      params.permit(:name, labels: {})
    end

    def heartbeat_params
      params.permit(:status, :currentRunId, labels: {})
    end
end
