# Production deployment

VELOHAM ships with PostgreSQL, Redis, Go API and Nginx frontend behind Caddy. Caddy obtains and renews Let's Encrypt certificates and redirects HTTP to HTTPS. The browser uses same-origin `/api/v1`, `/uploads`, and `/ws`; database, Redis and API ports remain private.

## Prerequisites

- A Linux host with Docker Engine and Docker Compose v2
- A DNS record pointing at the host
- Public TCP ports 80 and 443 and UDP port 443 allowed by the firewall
- Persistent-volume backups stored outside the host

## First deployment

```bash
cp .env.production.example .env.production
# Fill every required value with production secrets and the HTTPS public origin.
docker compose --env-file .env.production -f docker-compose.prod.yml config
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
docker compose --env-file .env.production -f docker-compose.prod.yml ps
curl --fail https://${DOMAIN}/healthz
```

Run the full public smoke check after DNS and TLS are ready:

```bash
./scripts/smoke.sh "https://${DOMAIN}"
```

The backend applies pending embedded SQL migrations before accepting traffic. Applied filenames are recorded in `schema_migrations`; never edit a migration already deployed. Add a new numbered migration instead.

## Upgrade

Back up PostgreSQL and uploaded files first, then build and replace containers:

```bash
./scripts/backup.sh ./backups
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
docker compose --env-file .env.production -f docker-compose.prod.yml ps
```

Keep the previous image tags until smoke tests pass. Database migrations are forward-only; restoring the pre-upgrade database backup is the rollback path for schema changes.

## Required operational controls

- Point the `DOMAIN` A/AAAA record to the server before starting Caddy.
- Back up the `postgres_data` and `uploads_data` volumes on a schedule; test restores.
- Restrict SSH and the public port with a firewall.
- Send container logs to retained centralized storage and alert on unhealthy containers or repeated 5xx responses.
- Rotate `JWT_SECRET`, database credentials, and administrator credentials through the deployment secret store.
- Narrow `TRUSTED_PROXIES` if the Docker network uses a more specific CIDR than the Compose default.
- Store user uploads in scanned object storage before scaling to multiple backend replicas.

## GitHub deployment

The manual `Deploy production` workflow deploys only `main`, creates a backup before rebuilding, and runs the public HTTPS smoke test. Configure the protected `production` environment with secrets `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_SSH_KEY`, pinned `DEPLOY_HOST_KEY`, `DEPLOY_PATH` and variable `PUBLIC_ORIGIN`. Keep `.env.production` only on the server.

## Restore drill

Stop application traffic before a restore, then pass one database dump and its matching uploads archive:

```bash
./scripts/restore.sh ./backups/database-TIMESTAMP.dump ./backups/uploads-TIMESTAMP.tar.gz
docker compose --env-file .env.production -f docker-compose.prod.yml restart backend frontend
curl --fail https://${DOMAIN}/healthz
```
