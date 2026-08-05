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

  FOREACH category_name IN ARRAY ARRAY[
    'Велосипеды целиком',
    'Рамы / фреймсеты',
    'Колёса',
    'Трансмиссия',
    'Руль и управление',
    'Тормоза',
    'Седло и посадка',
    'Аксессуары',
    'Экипировка',
    'Вело-услуги',
    'Вело-события'
  ]
  LOOP
    SELECT COUNT(*) INTO existing_count FROM listings WHERE category = category_name;
    need_count := GREATEST(0, 5 - existing_count);

    FOR i IN 1..need_count LOOP
      title_prefix := CASE category_name
        WHEN 'Велосипеды целиком' THEN 'Полный байк VELOHAM'
        WHEN 'Рамы / фреймсеты' THEN 'Фреймсет под сборку'
        WHEN 'Колёса' THEN 'Колёса / вилсет'
        WHEN 'Трансмиссия' THEN 'Трансмиссия и привод'
        WHEN 'Руль и управление' THEN 'Cockpit / управление'
        WHEN 'Тормоза' THEN 'Тормозной сет'
        WHEN 'Седло и посадка' THEN 'Седло и посадка'
        WHEN 'Аксессуары' THEN 'Аксессуар VELOHAM'
        WHEN 'Экипировка' THEN 'Экипировка rider'
        WHEN 'Вело-услуги' THEN 'Вело-услуга мастера'
        ELSE 'Вело-событие / покатушка'
      END;

      deal := CASE
        WHEN category_name IN ('Вело-услуги', 'Вело-события') THEN 'продажа'
        WHEN i % 3 = 0 THEN 'обмен'
        WHEN i % 2 = 0 THEN 'продажа или обмен'
        ELSE 'продажа'
      END;
      listing_id := uuid_generate_v4();

      INSERT INTO listings (
        id, user_id, title, description, price, city, category, condition, deal_type, status, created_at, updated_at
      ) VALUES (
        listing_id,
        seed_user_id,
        title_prefix || ' #' || (existing_count + i),
        CASE
          WHEN category_name = 'Вело-услуги' THEN 'Демо-услуга VELOHAM: ремонт, покраска, сборка, настройка или доставка.'
          WHEN category_name = 'Вело-события' THEN 'Демо-событие VELOHAM: покатушка, заезд, встреча или community ride.'
          ELSE 'Демо-объявление VELOHAM для категории ' || category_name || '.'
        END,
        CASE WHEN category_name = 'Вело-события' THEN 0 ELSE 5000 + ((existing_count + i) * 3200) END,
        CASE WHEN i % 4 = 0 THEN 'Ош' WHEN i % 3 = 0 THEN 'Каракол' ELSE 'Бишкек' END,
        category_name,
        CASE WHEN category_name IN ('Вело-услуги', 'Вело-события') THEN 'новое' WHEN i % 3 = 0 THEN 'хорошее' WHEN i % 2 = 0 THEN 'отличное' ELSE 'новое' END,
        deal,
        'active',
        now() - ((existing_count + i) || ' hours')::interval,
        now() - ((existing_count + i) || ' hours')::interval
      );

      IF category_name = 'Велосипеды целиком' THEN
        INSERT INTO build_cards (
          listing_id, frame, size, fork, wheels, hubs, tires, handlebar, stem, saddle,
          cranks, bottom_bracket, chain, cog, brakes, weight, frame_condition, defects, documents
        ) VALUES (
          listing_id, 'custom street/race frame', CASE WHEN i % 2 = 0 THEN 'M' ELSE 'L' END,
          'alloy/carbon fork', '700c wheelset', 'sealed hubs', 'urban tires',
          'riser/drop bar', '90mm stem', 'sport saddle', '48T crankset',
          'BSA bottom bracket', 'KMC chain', '17T cog / кассета',
          'caliper/disc brake', (8 + i)::text || ' кг', 'ровная, без трещин',
          CASE WHEN i % 3 = 0 THEN 'косметические следы' ELSE 'без дефектов' END,
          i % 2 = 0
        );
      END IF;

      IF deal <> 'продажа' THEN
        INSERT INTO match_preferences (
          listing_id, exchange_enabled, wants, categories, min_price, max_price,
          can_add_cash, max_cash_add, same_city_only
        ) VALUES (
          listing_id, true, 'Интересен обмен внутри VELOHAM',
          'Велосипеды целиком,Колёса,Рамы / фреймсеты',
          5000, 60000, i % 2 = 0, 10000, i % 3 = 0
        );
      END IF;
    END LOOP;
  END LOOP;
END $$;
