DO $$
DECLARE
  seed_user_id uuid;
  category_name text;
  existing_count integer;
  need_count integer;
  i integer;
  listing_id uuid;
  title_prefix text;
  deal text;
BEGIN
  INSERT INTO users (username, email, password_hash, city, contact, role, created_at)
  VALUES ('VELOHAM Demo', 'demo@veloham.kg', '$2a$10$demo.seed.user.password.hash', 'Бишкек', '@veloham_demo', 'user', now())
  ON CONFLICT (email) DO UPDATE SET username = EXCLUDED.username
  RETURNING id INTO seed_user_id;

  UPDATE users SET created_at = now() WHERE id = seed_user_id AND created_at IS NULL;

  FOREACH category_name IN ARRAY ARRAY['Fixed Gear', 'Road Bike', 'MTB', 'BMX', 'Frameset', 'Wheels', 'Parts', 'Accessories']
  LOOP
    SELECT COUNT(*) INTO existing_count
    FROM listings
    WHERE category = category_name;

    need_count := GREATEST(0, 5 - existing_count);

    FOR i IN 1..need_count LOOP
      title_prefix := CASE category_name
        WHEN 'Fixed Gear' THEN 'Fixed Gear street build'
        WHEN 'Road Bike' THEN 'Road Bike race setup'
        WHEN 'MTB' THEN 'MTB trail bike'
        WHEN 'BMX' THEN 'BMX street setup'
        WHEN 'Frameset' THEN 'Frameset для сборки'
        WHEN 'Wheels' THEN 'Wheelset / колеса'
        WHEN 'Parts' THEN 'Запчасти для апгрейда'
        ELSE 'Аксессуары VELOHAM'
      END;

      deal := CASE WHEN i % 3 = 0 THEN 'обмен' WHEN i % 2 = 0 THEN 'продажа или обмен' ELSE 'продажа' END;
      listing_id := uuid_generate_v4();

      INSERT INTO listings (
        id, user_id, title, description, price, city, category, condition, deal_type, status, created_at, updated_at
      ) VALUES (
        listing_id,
        seed_user_id,
        title_prefix || ' #' || (existing_count + i),
        'Демо-объявление VELOHAM для категории ' || category_name || '. Можно заменить на реальные данные, фото и описание.',
        8000 + ((existing_count + i) * 3500),
        CASE WHEN i % 4 = 0 THEN 'Ош' WHEN i % 3 = 0 THEN 'Каракол' ELSE 'Бишкек' END,
        category_name,
        CASE WHEN i % 4 = 0 THEN 'требует ремонта' WHEN i % 3 = 0 THEN 'хорошее' WHEN i % 2 = 0 THEN 'отличное' ELSE 'новое' END,
        deal,
        'active',
        now() - ((existing_count + i) || ' hours')::interval,
        now() - ((existing_count + i) || ' hours')::interval
      );

      IF category_name IN ('Fixed Gear', 'Road Bike', 'MTB', 'BMX') THEN
        INSERT INTO build_cards (
          listing_id, frame, size, fork, wheels, hubs, tires, handlebar, stem, saddle,
          cranks, bottom_bracket, chain, cog, brakes, weight, frame_condition, defects, documents
        ) VALUES (
          listing_id,
          category_name || ' frame',
          CASE WHEN i % 2 = 0 THEN 'M' ELSE 'L' END,
          'Carbon / alloy fork',
          '700c / 26 custom wheels',
          'Sealed bearing hubs',
          'Street tires',
          'Riser / drop bar',
          '90mm stem',
          'Sport saddle',
          '48T crankset',
          'BSA bottom bracket',
          'KMC chain',
          '17T cog / кассета',
          CASE WHEN category_name = 'Fixed Gear' THEN 'front brake optional' ELSE 'disc / caliper brakes' END,
          (8 + i)::text || ' кг',
          'ровная, без трещин',
          CASE WHEN i % 3 = 0 THEN 'есть косметические следы' ELSE 'без дефектов' END,
          i % 2 = 0
        );
      END IF;

      IF deal <> 'продажа' THEN
        INSERT INTO match_preferences (
          listing_id, exchange_enabled, wants, categories, min_price, max_price,
          can_add_cash, max_cash_add, same_city_only
        ) VALUES (
          listing_id,
          true,
          'Интересен обмен на велосипед или комплектующие',
          CASE category_name
            WHEN 'Fixed Gear' THEN 'Road Bike,Wheels'
            WHEN 'Road Bike' THEN 'Fixed Gear,MTB'
            WHEN 'MTB' THEN 'Road Bike,BMX'
            WHEN 'BMX' THEN 'Fixed Gear,Parts'
            ELSE 'Fixed Gear,Road Bike,MTB,BMX'
          END,
          5000,
          50000,
          i % 2 = 0,
          10000,
          i % 3 = 0
        );
      END IF;
    END LOOP;
  END LOOP;

  UPDATE listings
  SET created_at = now(), updated_at = now()
  WHERE user_id = seed_user_id AND (created_at IS NULL OR updated_at IS NULL);
END $$;
