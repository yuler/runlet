class CreateRunnerRuns < ActiveRecord::Migration[8.2]
  def change
    create_table :runner_runs, id: :uuid do |t|
      t.references :account, null: false, foreign_key: true, type: :uuid
      t.references :runner, null: false, foreign_key: true, type: :uuid
      t.references :identity, null: false, foreign_key: true, type: :uuid
      t.string :mode, null: false, default: "shell"
      t.string :status, null: false, default: "queued"
      t.text :command, null: false
      t.string :cwd
      t.json :env, null: false, default: {}
      t.integer :timeout_seconds
      t.integer :exit_code
      t.text :message
      t.datetime :claimed_at
      t.datetime :started_at
      t.datetime :finished_at

      t.timestamps
    end

    add_index :runner_runs, [ :runner_id, :status, :created_at ]
  end
end
