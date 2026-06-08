class DashboardsController < ApplicationController
  def show
    @sessions = Current.identity.sessions.order(created_at: :desc)
    @total_users = Current.account.users.count
    @total_sessions = @sessions.count
    @total_runners = Current.account.runners.count
  end
end
