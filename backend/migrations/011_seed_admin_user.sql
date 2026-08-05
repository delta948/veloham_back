INSERT INTO users (username, email, password_hash, city, contact, role, is_blocked, created_at)
VALUES (
  'VELOHAM Admin',
  'admin@veloham.kg',
  '$2a$10$tTDNbpvKzMbC93DyXFGCkOmkUq0QJ0taGw70IxebA0Ukc/.oR2Oja',
  'Бишкек',
  '@veloham_admin',
  'admin',
  false,
  now()
)
ON CONFLICT (email) DO UPDATE SET
  username = EXCLUDED.username,
  password_hash = EXCLUDED.password_hash,
  city = EXCLUDED.city,
  contact = EXCLUDED.contact,
  role = 'admin',
  is_blocked = false;
