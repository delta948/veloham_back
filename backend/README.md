# VELOHAM Backend

Backend маркетплейса VELOHAM на Go, Gin и PostgreSQL.

## Локальный запуск

```bash
cp .env.example .env
go run ./cmd/api
```

По умолчанию API доступен по адресу `http://127.0.0.1:8080/api/v1`, а проверка состояния — `http://127.0.0.1:8080/healthz`.

## Деплой на Render

1. Подключите этот репозиторий в Render через **New > Blueprint**.
2. Render автоматически прочитает корневой файл `render.yaml`.
3. Заполните секретные переменные окружения.

Основные адреса:

```env
CORS_ORIGIN=https://ВАШ_FRONTEND.vercel.app
API_BASE_URL=https://ВАШ_BACKEND.onrender.com
PUBLIC_BASE_URL=https://ВАШ_FRONTEND.vercel.app
```

- `CORS_ORIGIN` — точный HTTPS-адрес frontend без завершающего `/`.
- `API_BASE_URL` — адрес этого backend на Render без `/api/v1`.
- `PUBLIC_BASE_URL` — адрес frontend на Vercel.

Также production-режим требует настройки FreedomPay и SMTP. PostgreSQL, `JWT_SECRET` и постоянный диск для изображений создаются через Blueprint.

После деплоя проверьте `https://ВАШ_BACKEND.onrender.com/healthz`. Ответ должен быть `{"status":"ok"}`.
