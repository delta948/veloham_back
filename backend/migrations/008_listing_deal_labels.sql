ALTER TABLE listings
  ADD COLUMN IF NOT EXISTS labels jsonb NOT NULL DEFAULT '[]'::jsonb;

UPDATE listings
SET
  labels = (
    SELECT COALESCE(jsonb_agg(DISTINCT label), '[]'::jsonb)
    FROM (
      SELECT jsonb_array_elements_text(labels) AS label
      UNION ALL
      SELECT 'с моей доплатой' WHERE deal_type = 'обмен с моей доплатой'
      UNION ALL
      SELECT 'с вашей доплатой' WHERE deal_type = 'обмен с доплатой покупателя'
    ) AS source
    WHERE label IN ('срочно', 'торг', 'с моей доплатой', 'с вашей доплатой')
  ),
  deal_type = CASE
    WHEN deal_type IN ('обмен с моей доплатой', 'обмен с доплатой покупателя') THEN 'обмен'
    WHEN deal_type IN ('продажа', 'обмен', 'продажа или обмен') THEN deal_type
    ELSE 'продажа'
  END
WHERE deal_type NOT IN ('продажа', 'обмен', 'продажа или обмен')
   OR labels <> '[]'::jsonb;

UPDATE listings SET labels = '[]'::jsonb WHERE labels IS NULL;

CREATE INDEX IF NOT EXISTS idx_listings_deal_type ON listings (deal_type);
CREATE INDEX IF NOT EXISTS idx_listings_labels ON listings USING gin (labels);
