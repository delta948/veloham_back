ALTER TABLE listings
  ADD COLUMN IF NOT EXISTS bike_type varchar(40);

UPDATE listings
SET bike_type = CASE
  WHEN category = 'Fixed Gear' THEN 'fixed'
  WHEN category = 'Road Bike' THEN 'road'
  WHEN category = 'MTB' THEN 'mtb'
  WHEN category = 'BMX' THEN 'bmx'
  ELSE bike_type
END
WHERE bike_type IS NULL OR bike_type = '';

CREATE INDEX IF NOT EXISTS idx_listings_bike_type_frame_size ON listings (bike_type, frame_size);
