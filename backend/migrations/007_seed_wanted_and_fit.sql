DO $$
DECLARE
  seed_user_id uuid;
BEGIN
  SELECT id INTO seed_user_id FROM users WHERE email = 'demo@veloham.kg';

  WITH ranked AS (
    SELECT l.id, ROW_NUMBER() OVER (ORDER BY l.created_at DESC, l.id) rn
    FROM listings l
    WHERE l.user_id = seed_user_id
  )
  UPDATE listings l
  SET
    brand = (ARRAY['Cinelli','Specialized','Trek','Giant','Unknown','Canyon','Shimano','SRAM'])[((r.rn - 1) % 8) + 1],
    frame_size = (ARRAY['S','M','L','52','54','56','58'])[((r.rn - 1) % 7) + 1],
    rider_height_min = (ARRAY[155,165,172,168,175,180,185])[((r.rn - 1) % 7) + 1],
    rider_height_max = (ARRAY[168,176,185,178,188,192,198])[((r.rn - 1) % 7) + 1]
  FROM ranked r
  WHERE l.id = r.id;

  INSERT INTO wanted_requests (user_id, title, category, min_budget, max_budget, city, frame_size, rider_height, description, status)
  SELECT seed_user_id, title, category, min_budget, max_budget, city, frame_size, rider_height, description, 'active'
  FROM (VALUES
    ('Ищу fixed gear для города', 'Велосипеды целиком', 25000, 60000, 'Бишкек', 'M', 174, 'Нужен живой фикс или single speed, желательно без трещин и с нормальными колесами.'),
    ('Куплю вилсет 700c', 'Колёса', 8000, 30000, 'Ош', '54', 176, 'Ищу колеса под road/fixed, рассмотрю разные варианты.'),
    ('Нужна рама под сборку', 'Рамы / фреймсеты', 12000, 45000, 'Бишкек', 'L', 183, 'Интересует рама или frameset без критичных дефектов.'),
    ('Ищу тормоза для MTB', 'Тормоза', 3000, 18000, 'Каракол', '', 178, 'Гидравлика или хорошие механические тормоза.'),
    ('Куплю шлем и экипировку', 'Экипировка', 1000, 12000, 'Бишкек', '', 170, 'Нужен шлем, перчатки или очки для города.')
  ) AS v(title, category, min_budget, max_budget, city, frame_size, rider_height, description)
  WHERE NOT EXISTS (
    SELECT 1 FROM wanted_requests wr WHERE wr.user_id = seed_user_id AND wr.title = v.title
  );
END $$;
