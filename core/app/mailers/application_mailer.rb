class ApplicationMailer < ActionMailer::Base
  SUPPORT_EMAIL = ENV.fetch("SUPPORT_EMAIL", "support@example.com")
  default from: ENV.fetch("MAILER_FROM", "Runlet <#{SUPPORT_EMAIL}>")

  layout "mailer"
end
