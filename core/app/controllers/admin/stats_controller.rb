class Admin::StatsController < AdminController
  def show
    @accounts_count = Account.count
    @users_count = User.count

    accounts = Account.where(personal: true)

    user_counts = User.where(account_id: accounts.ids).group(:account_id).count

    @personal_accounts = accounts.map do |account|
      {
        name: account.name,
        users: user_counts[account.id] || 0,
        created_at: account.created_at
      }
    end.sort_by { |a| -a[:users] }
  end
end
