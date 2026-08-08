# Production deployment

VELOHAM ships with PostgreSQL, Redis, Go API and a containerized frontend behind Nginx installed on the server. Docker publishes the frontend only on `127.0.0.1:8080`; Nginx provides the public domain and HTTPS. The browser uses same-origin `/api/v1`, `/uploads`, and `/ws`; database, Redis and API ports remain private.

## Prerequisites

- A Linux host with Docker Engine and Docker Compose v2
- A DNS record pointing at the host
- Nginx and Certbot installed on the server
- Public TCP ports 80 and 443 allowed by the firewall
- Persistent-volume backups stored outside the host

## First deployment

```bash
cp .env.production.example .env.production
# Fill every required value with production secrets and the HTTPS public origin.
docker compose --env-file .env.production -f docker-compose.prod.yml config
docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build
docker compose --env-file .env.production -f docker-compose.prod.yml ps
curl --fail http://127.0.0.1:${HTTP_PORT:-8080}/healthz
```

## Upload through WinSCP

1. Connect to the server over SFTP in WinSCP.
2. Upload the project to a dedicated directory such as `/opt/veloham`. Do not upload local `.env` files, `frontend/node_modules`, `frontend/dist`, backups, or `.git` credentials.
3. Copy `.env.production.example` to `.env.production` on the server and fill in the production values there.
4. Open the WinSCP terminal in `/opt/veloham` and run the Compose commands shown above.

For later updates, upload changed project files and run `./scripts/backup.sh ./backups` before `docker compose --env-file .env.production -f docker-compose.prod.yml up -d --build`.

## Server Nginx and HTTPS

Upload `deploy/nginx/veloham.conf.example` to the server, replace every `YOUR_DOMAIN` with the real domain, and install it as `/etc/nginx/sites-available/veloham`. The template starts over HTTP so Nginx can pass its first configuration check; Certbot then adds HTTPS and the redirect automatically.

```bash
sudo ln -s /etc/nginx/sites-available/veloham /etc/nginx/sites-enabled/veloham
sudo nginx -t
sudo systemctl reload nginx
sudo certbot --nginx -d example.kg -d www.example.kg
sudo nginx -t
sudo systemctl reload nginx
```

The upstream must remain `http://127.0.0.1:8080`. Do not expose PostgreSQL, Redis, or backend port 8080 publicly.

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

- Point the domain A/AAAA record to the server before requesting the certificate.
- Back up the `postgres_data` and `uploads_data` volumes on a schedule; test restores.
- Restrict SSH and the public port with a firewall.
- Send container logs to retained centralized storage and alert on unhealthy containers or repeated 5xx responses.
- Rotate `JWT_SECRET`, database credentials, and administrator credentials through the deployment secret store.
- Narrow `TRUSTED_PROXIES` if the Docker network uses a more specific CIDR than the Compose default.
- Store user uploads in scanned object storage before scaling to multiple backend replicas.

## GitHub deployment

If you later replace WinSCP updates with Git, the manual `Deploy production` workflow deploys only `main`, creates a backup before rebuilding, and runs the public HTTPS smoke test. Configure the protected `production` environment with secrets `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_SSH_KEY`, pinned `DEPLOY_HOST_KEY`, `DEPLOY_PATH` and variable `PUBLIC_ORIGIN`. Keep `.env.production` only on the server.

## Restore drill

Stop application traffic before a restore, then pass one database dump and its matching uploads archive:

```bash
./scripts/restore.sh ./backups/database-TIMESTAMP.dump ./backups/uploads-TIMESTAMP.tar.gz
docker compose --env-file .env.production -f docker-compose.prod.yml restart backend frontend
curl --fail https://${DOMAIN}/healthz
```
