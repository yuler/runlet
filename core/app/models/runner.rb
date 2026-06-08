class Runner < ApplicationRecord
  STATUSES = %w[idle running offline].freeze

  belongs_to :account
  belongs_to :identity

  validates :name, presence: true, length: { maximum: 120 }
  validates :status, presence: true, inclusion: { in: STATUSES }
  validate :labels_must_be_a_flat_string_hash

  normalizes :name, with: ->(value) { value.to_s.strip.presence }

  scope :for_account, ->(account) { where(account:) }

  def heartbeat!(status:, current_run_id:, labels:)
    update!(
      status: status.presence || "idle",
      current_run_id: current_run_id.presence,
      labels: self.class.normalize_labels(labels),
      last_heartbeat_at: Time.current
    )
  end

  def self.register!(account:, identity:, name:, labels:)
    runner = find_or_initialize_by(account:, name: name.to_s.strip)
    runner.identity = identity
    runner.labels = normalize_labels(labels)
    runner.status ||= "idle"
    runner.last_heartbeat_at ||= Time.current
    runner.save!
    runner
  end

  def self.normalize_labels(labels)
    (labels || {}).to_h.each_with_object({}) do |(key, value), normalized|
      normalized[key.to_s] = value.to_s
    end
  end

  private
    def labels_must_be_a_flat_string_hash
      return if labels.is_a?(Hash) && labels.all? { |key, value| key.is_a?(String) && value.is_a?(String) }

      errors.add(:labels, "must be a flat object with string values")
    end
end
