class Api::V1::Runners::ClaimsController < Api::V1::BaseController
  before_action :set_runner

  def create
    render json: { run: nil }
  end

  private
    def set_runner
      @runner = Current.account.runners.find(params[:runner_id])
    end
end
