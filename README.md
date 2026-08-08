# VELOHAM

[![CI](https://github.com/delta948/veloham/actions/workflows/ci.yml/badge.svg)](https://github.com/delta948/veloham/actions/workflows/ci.yml)

VELOHAM is a two-service marketplace for bikes, fixed gear, road, MTB, BMX, parts and accessories.

Security policy and the latest audit status: [SECURITY.md](SECURITY.md)

## Services

- `backend`: Go + Gin + PostgreSQL + JWT + WebSocket + uploads
- `frontend`: React + TypeScript + Vite + Tailwind + React Router + Axios + Zustand

## Backend Architecture

Backend is now prepared as a modular monolith:

- new API namespace: `http://127.0.0.1:8080/api/v1`
- legacy API namespace: `http://127.0.0.1:8080/api`
- module packages live in `backend/internal/modules`
- shared infrastructure lives in `backend/internal/common`
- architecture notes: `backend/ARCHITECTURE.md`

The project still runs as one Go process, but modules are registered independently so chat, search, notifications, auth, or buy requests can later be extracted into separate services.

## Run

```bash
docker compose up --build
```

Frontend: http://127.0.0.1:5173  
Backend API: http://127.0.0.1:8080/api/v1

Local backend without Docker:

```bash
cd backend
cp .env.example .env
go run ./cmd/api
```

Local frontend:

```bash
cd frontend
npm install
npm run dev
```

## Production security

- Set `APP_ENV=production` and provide a unique `JWT_SECRET` with at least 32 characters.
- Set explicit HTTPS origins in `CORS_ORIGIN`; wildcards are rejected in production.
- To provision an administrator, set both `ADMIN_EMAIL` and `ADMIN_PASSWORD`. The password must contain at least 6 characters.
- New registrations require a one-time code sent by email. Configure `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, and `SMTP_FROM`; production refuses to start without them.
- Do not commit `.env` files or production credentials.

Production Compose configuration, migration behavior, backups, upgrades, and rollback are documented in [DEPLOYMENT.md](DEPLOYMENT.md).

Production includes automatic HTTPS through Caddy. A manual protected GitHub deployment workflow performs a backup, fast-forward update, rebuild, and public smoke test.
