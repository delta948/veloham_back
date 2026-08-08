# Production deployment

VELOHAM ships as three private services (PostgreSQL, Redis, Go API) behind one public Nginx frontend. The browser uses same-origin `/api/v1`, `/uploads`, and `/ws` routes, so production does not expose the database, Redis, or API container ports.

## Prerequisites

- A Linux host with Docker Engine and Docker Compose v2
- A DNS record pointing at the host
- TLS termination in a host reverse proxy or load balancer
- Persistent-volume backups stored outside the host

## First deployment

```bash
cp .env.production.example .env.production
# Fill every required value with production secrets and the HTTPS public origin.
docker compose --env-file .env.production -f docker-compose.prod.yml config
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
docker compose --env-file .env.production -f docker-compose.prod.yml ps
curl --fail http://127.0.0.1:${HTTP_PORT:-80}/healthz
```

The backend applies pending embedded SQL migrations before accepting traffic. Applied filenames are recorded in `schema_migrations`; never edit a migration already deployed. Add a new numbered migration instead.

## Upgrade

Back up PostgreSQL and uploaded files first, then build and replace containers:

```bash
docker compose --env-file .env.production -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U veloham -d veloham -Fc > veloham-before-upgrade.dump
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
docker compose --env-file .env.production -f docker-compose.prod.yml ps
```

Keep the previous image tags until smoke tests pass. Database migrations are forward-only; restoring the pre-upgrade database backup is the rollback path for schema changes.

## Required operational controls

- Terminate HTTPS and redirect HTTP to HTTPS.
- Back up the `postgres_data` and `uploads_data` volumes on a schedule; test restores.
- Restrict SSH and the public port with a firewall.
- Send container logs to retained centralized storage and alert on unhealthy containers or repeated 5xx responses.
- Rotate `JWT_SECRET`, database credentials, and administrator credentials through the deployment secret store.
- Narrow `TRUSTED_PROXIES` if the Docker network uses a more specific CIDR than the Compose default.
- Store user uploads in scanned object storage before scaling to multiple backend replicas.
