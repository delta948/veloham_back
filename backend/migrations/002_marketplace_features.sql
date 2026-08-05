ALTER TABLE users
  ADD COLUMN IF NOT EXISTS city varchar(120),
  ADD COLUMN IF NOT EXISTS contact varchar(160),
  ADD COLUMN IF NOT EXISTS role varchar(30) NOT NULL DEFAULT 'user',
  ADD COLUMN IF NOT EXISTS is_blocked boolean NOT NULL DEFAULT false;

ALTER TABLE listings
  ADD COLUMN IF NOT EXISTS deal_type varchar(80) NOT NULL DEFAULT 'продажа',
  ADD COLUMN IF NOT EXISTS views integer NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS build_cards (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  listing_id uuid UNIQUE NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
  frame text,
  size text,
  fork text,
  wheels text,
  hubs text,
  tires text,
  handlebar text,
  stem text,
  saddle text,
  cranks text,
  bottom_bracket text,
  chain text,
  cog text,
  brakes text,
  weight text,
  frame_condition text,
  defects text,
  documents boolean NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS match_preferences (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  listing_id uuid UNIQUE NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
  exchange_enabled boolean NOT NULL DEFAULT false,
  wants text,
  categories text,
  min_price integer NOT NULL DEFAULT 0,
  max_price integer NOT NULL DEFAULT 0,
  can_add_cash boolean NOT NULL DEFAULT false,
  max_cash_add integer NOT NULL DEFAULT 0,
  same_city_only boolean NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS reviews (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  seller_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  author_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  listing_id uuid REFERENCES listings(id) ON DELETE SET NULL,
  rating integer NOT NULL CHECK (rating >= 1 AND rating <= 5),
  text text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS reports (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  reporter_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  listing_id uuid REFERENCES listings(id) ON DELETE SET NULL,
  seller_id uuid REFERENCES users(id) ON DELETE SET NULL,
  reason varchar(80) NOT NULL,
  text text,
  status varchar(30) NOT NULL DEFAULT 'new',
  created_at timestamptz NOT NULL DEFAULT now()
);
