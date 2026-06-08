class RunnerRunEvent < ApplicationRecord
  LEVELS = %w[debug info warn error].freeze
  STREAMS = %w[runner stdout stderr].freeze

  belongs_to :runner_run

  validates :sequence, presence: true, uniqueness: { scope: :runner_run_id }
  validates :level, presence: true, inclusion: { in: LEVELS }
  validates :stream, allow_blank: true, inclusion: { in: STREAMS }
  validates :message, presence: true
  validates :occurred_at, presence: true
end
