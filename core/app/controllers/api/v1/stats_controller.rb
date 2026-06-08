class Api::V1::StatsController < Api::V1::BaseController
  def show
    render json: {
      users: Current.account.users.count,
      sessions: Current.identity.sessions.count
    }
  end
end
