DO $$
BEGIN
  WITH ranked AS (
    SELECT
      l.id,
      l.category,
      ROW_NUMBER() OVER (PARTITION BY l.category ORDER BY l.created_at DESC, l.id) AS rn
    FROM listings l
    JOIN users u ON u.id = l.user_id
    WHERE u.email = 'demo@veloham.kg'
  ),
  prepared AS (
    SELECT
      id,
      category,
      rn,
      (ARRAY[
        'ночной сетап',
        'городской апгрейд',
        'чистая сборка',
        'боевой комплект',
        'лайтовый вариант'
      ])[((rn - 1) % 5) + 1] AS variant,
      (ARRAY[
        'Бишкек',
        'Ош',
        'Каракол',
        'Чолпон-Ата',
        'Джалал-Абад'
      ])[((rn - 1) % 5) + 1] AS demo_city,
      (ARRAY[
        'новое',
        'отличное',
        'хорошее',
        'отличное',
        'требует ремонта'
      ])[((rn - 1) % 5) + 1] AS demo_condition,
      (ARRAY[
        'продажа',
        'продажа или обмен',
        'обмен',
        'продажа',
        'обмен с доплатой покупателя'
      ])[((rn - 1) % 5) + 1] AS demo_deal
    FROM ranked
  )
  UPDATE listings l
  SET
    title = CASE p.category
      WHEN 'Велосипеды целиком' THEN (ARRAY['Cinelli style fixed build','Aero road Бишкек','Trail hardtail setup','BMX park complete','Urban commuter custom'])[p.rn]
      WHEN 'Рамы / фреймсеты' THEN (ARRAY['Steel track frameset','Aluminum gravel frame','Carbon road frameset','BMX frame raw finish','MTB hardtail frame'])[p.rn]
      WHEN 'Колёса' THEN (ARRAY['Deep rim wheelset 60mm','MTB 29 wheelset','Fixed flip-flop wheels','Road clincher wheels','BMX double wall wheels'])[p.rn]
      WHEN 'Трансмиссия' THEN (ARRAY['SRAM drivetrain kit','Shimano 105 groupset','Fixed 48T crank combo','MTB cassette + derailleur','Chain and cog pack'])[p.rn]
      WHEN 'Руль и управление' THEN (ARRAY['Wide riser cockpit','Drop bar + stem set','BMX handlebar chrome','Bullhorn fixed setup','Carbon road cockpit'])[p.rn]
      WHEN 'Тормоза' THEN (ARRAY['Hydraulic disc brakes','Road caliper pair','BMX U-brake kit','MTB rotor and pads','Fixed front brake clean'])[p.rn]
      WHEN 'Седло и посадка' THEN (ARRAY['Selle style saddle','Aero seatpost setup','Comfort city saddle','BMX pivotal seat','Bike fit cockpit pack'])[p.rn]
      WHEN 'Аксессуары' THEN (ARRAY['Kryptonite lock set','Night light pack','Frame bag street','Bottle cage duo','Tool roll compact'])[p.rn]
      WHEN 'Экипировка' THEN (ARRAY['Aero helmet black','Street gloves pair','Cycling glasses neon','Rain jacket ride','Knee pads BMX'])[p.rn]
      WHEN 'Вело-услуги' THEN (ARRAY['Полная сборка велосипеда','Покраска рамы порошком','Настройка трансмиссии','Доставка велосипеда по городу','Диагностика перед покупкой'])[p.rn]
      WHEN 'Вело-события' THEN (ARRAY['Friday fixed ride','MTB Каракол weekend','BMX jam session','Road sunrise заезд','Community night cruise'])[p.rn]
      WHEN 'Fixed Gear' THEN (ARRAY['NJS-inspired fixed gear','Chrome street fixed','Daily brakeless setup','Tracklocross fixed build','Minimal black fixie'])[p.rn]
      WHEN 'Road Bike' THEN (ARRAY['Carbon road rocket','Endurance road bike','Aero sprint machine','Alu road training bike','Classic steel road'])[p.rn]
      WHEN 'MTB' THEN (ARRAY['Trail hardtail 29','Enduro-ready bike','XC light build','Mountain bomber','Karakol trail setup'])[p.rn]
      WHEN 'BMX' THEN (ARRAY['Street BMX complete','Park BMX clean','Raw frame BMX','BMX for tricks','Dirt jump style BMX'])[p.rn]
      WHEN 'Frameset' THEN (ARRAY['Track frameset steel','Road frameset carbon','BMX frame black','MTB frame boost','Gravel frameset'])[p.rn]
      WHEN 'Wheels' THEN (ARRAY['Fixed deep wheels','Road wheelset fast','MTB 29 wheels','BMX wheel pair','Gravel wheelset'])[p.rn]
      WHEN 'Parts' THEN (ARRAY['Crankset and chain','Handlebar set','Brake upgrade','Pedals and straps','Service parts pack'])[p.rn]
      WHEN 'Accessories' THEN (ARRAY['Lock and lights','Bike bag set','Tool kit compact','Bottle setup','City accessory pack'])[p.rn]
      ELSE p.category || ' — ' || p.variant
    END,
    description = CASE p.category
      WHEN 'Вело-услуги' THEN 'Услуга от demo-мастера VELOHAM: реальные сроки, понятная цена, можно написать в чат и обсудить детали.'
      WHEN 'Вело-события' THEN 'Событие VELOHAM community: встреча райдеров, маршрут, темп и детали обсуждаются в чате.'
      ELSE 'Разное demo-объявление VELOHAM: ' || p.variant || '. Состояние, комплектация и обмен отличаются от других карточек.'
    END,
    city = p.demo_city,
    condition = p.demo_condition,
    deal_type = p.demo_deal,
    price = CASE
      WHEN p.category = 'Вело-события' THEN 0
      WHEN p.category = 'Вело-услуги' THEN 1500 + p.rn * 1800
      WHEN p.category IN ('Аксессуары', 'Экипировка') THEN 900 + p.rn * 1600
      WHEN p.category IN ('Тормоза', 'Седло и посадка', 'Руль и управление', 'Трансмиссия') THEN 2500 + p.rn * 3200
      WHEN p.category IN ('Колёса', 'Wheels') THEN 9000 + p.rn * 6500
      WHEN p.category IN ('Рамы / фреймсеты', 'Frameset') THEN 12000 + p.rn * 8500
      ELSE 18000 + p.rn * 12000
    END
  FROM prepared p
  WHERE l.id = p.id;

  DELETE FROM listing_images
  WHERE listing_id IN (
    SELECT l.id
    FROM listings l
    JOIN users u ON u.id = l.user_id
    WHERE u.email = 'demo@veloham.kg'
  )
  AND image_url LIKE 'https://images.unsplash.com/%';

  WITH ranked AS (
    SELECT
      l.id,
      l.category,
      ROW_NUMBER() OVER (PARTITION BY l.category ORDER BY l.created_at DESC, l.id) AS rn
    FROM listings l
    JOIN users u ON u.id = l.user_id
    WHERE u.email = 'demo@veloham.kg'
  ),
  images AS (
    SELECT
      id,
      (ARRAY[
        'https://images.unsplash.com/photo-1485965120184-e220f721d03e?q=80&w=1200&auto=format&fit=crop',
        'https://images.unsplash.com/photo-1507035895480-2b3156c31fc8?q=80&w=1200&auto=format&fit=crop',
        'https://images.unsplash.com/photo-1571068316344-75bc76f77890?q=80&w=1200&auto=format&fit=crop',
        'https://images.unsplash.com/photo-1511994298241-608e28f14fde?q=80&w=1200&auto=format&fit=crop',
        'https://images.unsplash.com/photo-1529422643029-d4585747aaf2?q=80&w=1200&auto=format&fit=crop',
        'https://images.unsplash.com/photo-1496147433903-1e62fdb6f2be?q=80&w=1200&auto=format&fit=crop',
        'https://images.unsplash.com/photo-1558611848-73f7eb4001a1?q=80&w=1200&auto=format&fit=crop',
        'https://images.unsplash.com/photo-1544191696-102dbdaeeaa5?q=80&w=1200&auto=format&fit=crop',
        'https://images.unsplash.com/photo-1541625602330-2277a4c46182?q=80&w=1200&auto=format&fit=crop',
        'https://images.unsplash.com/photo-1517649763962-0c623066013b?q=80&w=1200&auto=format&fit=crop'
      ])[((ABS(HASHTEXT(category)) + rn) % 10) + 1] AS image_url
    FROM ranked
  )
  INSERT INTO listing_images (listing_id, image_url, sort_order)
  SELECT id, image_url, 0
  FROM images;
END $$;
