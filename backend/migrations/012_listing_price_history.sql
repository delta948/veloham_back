ALTER TABLE listings ADD COLUMN IF NOT EXISTS initial_price integer;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'listings_price_positive') THEN
    ALTER TABLE listings ADD CONSTRAINT listings_price_positive CHECK (price >= 0);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'listings_initial_price_positive') THEN
    ALTER TABLE listings ADD CONSTRAINT listings_initial_price_positive CHECK (initial_price IS NULL OR initial_price >= 0);
  END IF;
END $$;

UPDATE listings SET initial_price = price WHERE initial_price IS NULL;
ALTER TABLE listings ALTER COLUMN initial_price SET NOT NULL;

CREATE TABLE IF NOT EXISTS listing_price_history (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  listing_id uuid NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
  old_price integer NOT NULL CHECK (old_price > 0),
  new_price integer NOT NULL CHECK (new_price > 0),
  changed_at timestamptz NOT NULL DEFAULT now(),
  changed_by uuid NOT NULL REFERENCES users(id),
  ip_address inet,
  suspicious boolean NOT NULL DEFAULT false,
  suspicious_reason text,
  CONSTRAINT listing_price_actually_changed CHECK (old_price <> new_price)
);

CREATE INDEX IF NOT EXISTS idx_price_history_listing_changed ON listing_price_history(listing_id, changed_at DESC);

CREATE TABLE IF NOT EXISTS notifications (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  listing_id uuid REFERENCES listings(id) ON DELETE CASCADE,
  price_history_id uuid REFERENCES listing_price_history(id) ON DELETE CASCADE,
  type varchar(40) NOT NULL,
  message text NOT NULL,
  link text NOT NULL,
  is_read boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, price_history_id)
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications(user_id, created_at DESC);

-- История является аудит-логом: API удаления/редактирования отсутствует.
-- Дополнительно запрещаем UPDATE даже при случайном вызове из приложения.
CREATE OR REPLACE FUNCTION reject_price_history_mutation() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'listing price history is immutable';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS listing_price_history_immutable ON listing_price_history;
CREATE TRIGGER listing_price_history_immutable
BEFORE UPDATE ON listing_price_history
FOR EACH ROW EXECUTE FUNCTION reject_price_history_mutation();
