# VELOHAM

VELOHAM is a two-service marketplace for bikes, fixed gear, road, MTB, BMX, parts and accessories.

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
Backend API: http://127.0.0.1:8080/api

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
