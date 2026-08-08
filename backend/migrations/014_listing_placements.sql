CREATE TABLE IF NOT EXISTS listing_placements (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  listing_id uuid UNIQUE REFERENCES listings(id) ON DELETE SET NULL,
  kind varchar(20) NOT NULL CHECK (kind IN ('free', 'paid')),
  target_status varchar(30) NOT NULL DEFAULT 'active',
  amount integer NOT NULL DEFAULT 0 CHECK (amount >= 0),
  currency varchar(3) NOT NULL DEFAULT 'KGS',
  status varchar(20) NOT NULL CHECK (status IN ('pending', 'paid', 'failed')),
  provider varchar(30),
  provider_payment_id varchar(100) UNIQUE,
  checkout_url text,
  created_at timestamptz NOT NULL DEFAULT now(),
  paid_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_listing_placements_user_kind_status
  ON listing_placements(user_id, kind, status);

-- Preserve the lifetime free allowance for existing sellers, even if a listing
-- is deleted later. Only their first three historical listings consume it.
WITH ranked AS (
  SELECT id, user_id, status, created_at,
         ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at, id) AS position
  FROM listings
)
INSERT INTO listing_placements (user_id, listing_id, kind, target_status, amount, status, paid_at, created_at)
SELECT user_id, id, 'free', status, 0, 'paid', created_at, created_at
FROM ranked
WHERE position <= 3
ON CONFLICT (listing_id) DO NOTHING;
