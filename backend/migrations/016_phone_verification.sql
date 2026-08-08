ALTER TABLE users
  ADD COLUMN IF NOT EXISTS phone varchar(13);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique
  ON users(phone)
  WHERE phone IS NOT NULL AND phone <> '';

CREATE TABLE IF NOT EXISTS pending_registrations (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  username varchar(80) NOT NULL,
  email varchar(160) UNIQUE NOT NULL,
  phone varchar(13) UNIQUE NOT NULL,
  city varchar(120) NOT NULL DEFAULT '',
  contact varchar(160) NOT NULL DEFAULT '',
  password_hash text NOT NULL,
  provider_token text NOT NULL,
  code_hash text NOT NULL DEFAULT '',
  attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  resend_count integer NOT NULL DEFAULT 0 CHECK (resend_count >= 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  resend_after timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pending_registrations_expires_at
  ON pending_registrations(expires_at);
