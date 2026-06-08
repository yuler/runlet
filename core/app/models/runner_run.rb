class RunnerRun < ApplicationRecord
  MODES = %w[shell].freeze
  STATUSES = %w[queued running succeeded failed timed_out canceled].freeze

  belongs_to :account
  belongs_to :runner
  belongs_to :identity
  has_many :events, class_name: "RunnerRunEvent", dependent: :destroy

  validates :mode, presence: true, inclusion: { in: MODES }
  validates :status, presence: true, inclusion: { in: STATUSES }
  validates :command, presence: true
  validate :env_must_be_a_flat_string_hash

  scope :recent, -> { order(created_at: :desc) }
  scope :claimable, -> { where(status: "queued").order(created_at: :asc) }

  def claim!(runner)
    with_lock do
      return nil unless queued?
      return nil unless self.runner == runner

      update!(
        status: "running",
        claimed_at: Time.current,
        started_at: Time.current
      )
      self
    end
  end

  def record_event!(sequence:, level:, stream:, message:, metadata:, occurred_at:)
    events.create!(
      sequence:,
      level: level.presence || "info",
      stream: stream.presence,
      message:,
      metadata: self.class.normalize_env(metadata),
      occurred_at: occurred_at || Time.current
    )
  end

  def finish!(status:, exit_code:, finished_at:, message:)
    update!(
      status:,
      exit_code:,
      finished_at: finished_at || Time.current,
      message: message.presence
    )
  end

  def queued?
    status == "queued"
  end

  def self.normalize_env(env)
    (env || {}).to_h.each_with_object({}) do |(key, value), normalized|
      normalized[key.to_s] = value.to_s
    end
  end

  private
    def env_must_be_a_flat_string_hash
      return if env.is_a?(Hash) && env.all? { |key, value| key.is_a?(String) && value.is_a?(String) }

      errors.add(:env, "must be a flat object with string values")
    end
end
