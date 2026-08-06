UPDATE users
SET role = 'user',
    is_blocked = true,
    password_hash = 'disabled-known-credential'
WHERE email = 'admin@veloham.kg';
