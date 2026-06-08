require "test_helper"

class Api::V1::StatsControllerTest < ActionDispatch::IntegrationTest
  setup do
    @identity = Identity.create!(email: "test@#{SecureRandom.hex}.com")
    @account = Account.create!(name: "Test Account", personal: true)
    @user = User.create!(identity: @identity, account: @account, name: "Test User", role: "owner")
    @session = @identity.sessions.create!(user_agent: "TestAgent", ip_address: "127.0.0.1")
  end

  test "should get stats when authenticated" do
    get api_v1_stats_url, headers: { "Authorization" => "Bearer #{@session.signed_id}" }, as: :json
    assert_response :success
    json = JSON.parse(response.body)
    assert_equal 1, json["users"]
    assert_equal 1, json["sessions"]
  end

  test "should return unauthorized if not authenticated" do
    get api_v1_stats_url, as: :json
    assert_response :unauthorized
  end
end
