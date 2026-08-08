CREATE TABLE IF NOT EXISTS user_block_events (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  admin_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  action varchar(20) NOT NULL CHECK (action IN ('blocked', 'unblocked')),
  reason text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_block_events_user_created
  ON user_block_events(user_id, created_at DESC);
