ALTER TABLE listings
  ADD COLUMN IF NOT EXISTS frame_size_text varchar(40),
  ADD COLUMN IF NOT EXISTS recommended_height_min integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS recommended_height_max integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS is_urgent boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS is_bargain boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS is_exchange boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS extra_payment_from_me boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS extra_payment_from_buyer boolean NOT NULL DEFAULT false;

UPDATE listings
SET
  frame_size_text = COALESCE(NULLIF(frame_size_text, ''), frame_size),
  recommended_height_min = CASE WHEN recommended_height_min = 0 THEN rider_height_min ELSE recommended_height_min END,
  recommended_height_max = CASE WHEN recommended_height_max = 0 THEN rider_height_max ELSE recommended_height_max END,
  is_urgent = is_urgent OR labels @> '["срочно"]'::jsonb,
  is_bargain = is_bargain OR labels @> '["торг"]'::jsonb,
  is_exchange = is_exchange OR deal_type IN ('обмен', 'продажа или обмен') OR labels @> '["обмен"]'::jsonb,
  extra_payment_from_me = extra_payment_from_me OR labels @> '["с моей доплатой"]'::jsonb,
  extra_payment_from_buyer = extra_payment_from_buyer OR labels @> '["с вашей доплатой"]'::jsonb;

ALTER TABLE wanted_requests
  ADD COLUMN IF NOT EXISTS budget_min integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS budget_max integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS height integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS preferred_bike_type varchar(40);

UPDATE wanted_requests
SET
  budget_min = CASE WHEN budget_min = 0 THEN min_budget ELSE budget_min END,
  budget_max = CASE WHEN budget_max = 0 THEN max_budget ELSE budget_max END,
  height = CASE WHEN height = 0 THEN rider_height ELSE height END;

CREATE INDEX IF NOT EXISTS idx_listings_fit_contract ON listings (category, bike_type, frame_size_text, recommended_height_min, recommended_height_max);
CREATE INDEX IF NOT EXISTS idx_listings_boolean_labels ON listings (is_urgent, is_bargain, is_exchange, extra_payment_from_me, extra_payment_from_buyer);
CREATE INDEX IF NOT EXISTS idx_wanted_contract_filters ON wanted_requests (category, city, preferred_bike_type, budget_min, budget_max, height);
