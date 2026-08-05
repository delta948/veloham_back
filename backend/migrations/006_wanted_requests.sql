ALTER TABLE listings
  ADD COLUMN IF NOT EXISTS brand varchar(120),
  ADD COLUMN IF NOT EXISTS frame_size varchar(40),
  ADD COLUMN IF NOT EXISTS rider_height_min integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS rider_height_max integer NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS wanted_requests (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title varchar(180) NOT NULL,
  category varchar(80) NOT NULL,
  min_budget integer NOT NULL DEFAULT 0,
  max_budget integer NOT NULL DEFAULT 0,
  city varchar(120) NOT NULL,
  frame_size varchar(40),
  rider_height integer NOT NULL DEFAULT 0,
  description text,
  status varchar(30) NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wanted_offers (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  wanted_id uuid NOT NULL REFERENCES wanted_requests(id) ON DELETE CASCADE,
  seller_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  listing_id uuid NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
  message text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_wanted_requests_filters ON wanted_requests (category, city, status, min_budget, max_budget, frame_size);
CREATE INDEX IF NOT EXISTS idx_listings_fit_filters ON listings (brand, frame_size, rider_height_min, rider_height_max);
