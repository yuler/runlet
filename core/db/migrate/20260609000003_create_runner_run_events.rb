class CreateRunnerRunEvents < ActiveRecord::Migration[8.2]
  def change
    create_table :runner_run_events, id: :uuid do |t|
      t.references :runner_run, null: false, foreign_key: true, type: :uuid
      t.integer :sequence, null: false
      t.string :level, null: false, default: "info"
      t.string :stream
      t.text :message, null: false
      t.json :metadata, null: false, default: {}
      t.datetime :occurred_at, null: false

      t.timestamps
    end

    add_index :runner_run_events, [ :runner_run_id, :sequence ], unique: true
  end
end
