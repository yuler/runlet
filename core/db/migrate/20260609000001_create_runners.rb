class CreateRunners < ActiveRecord::Migration[8.2]
  def change
    create_table :runners, id: :uuid do |t|
      t.references :account, null: false, foreign_key: false, type: :uuid
      t.references :identity, null: false, foreign_key: false, type: :uuid
      t.string :name, null: false
      t.string :status, null: false, default: "idle"
      t.string :current_run_id
      t.json :labels, null: false, default: {}
      t.datetime :last_heartbeat_at

      t.timestamps
    end

    add_index :runners, [ :account_id, :name ], unique: true
    add_index :runners, [ :account_id, :status ]
    add_index :runners, :last_heartbeat_at
  end
end
