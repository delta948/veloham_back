# Security

## Reporting vulnerabilities

Do not publish credentials or exploit details in a public issue. Contact the repository owner privately and include the affected endpoint, reproduction steps, and expected impact.

## Security baseline

- Production requires a unique JWT secret of at least 32 characters and explicit CORS origins.
- Administrator provisioning is opt-in through `ADMIN_EMAIL` and `ADMIN_PASSWORD`; no privileged account is seeded from source control.
- Passwords must contain 6 to 72 characters and are hashed with bcrypt.
- Authentication endpoints are rate-limited per client IP.
- Existing tokens are rejected when their user is missing or blocked.
- JWT verification requires HS256, an expiration claim, and valid timestamps.
- WebSocket connections validate their Origin and receive the token through a subprotocol header instead of the URL.
- Uploads accept at most eight JPEG, PNG, or WebP images per request, with a 5 MB limit per file and generated server-side filenames.
- Public API responses do not expose account email addresses.

## Audit status (2026-08-07)

The backend passes `go test ./...`, `go vet ./...`, and `govulncheck` with no reachable vulnerabilities. CI runs PostgreSQL-backed migration and authorization integration tests. The frontend production build passes.

`npm audit` currently reports a high-severity React Router advisory affecting server-side RSC action handling. VELOHAM is a client-only Vite SPA and does not use React Router RSC, server actions, or SSR. The advisory's fixed 8.3.0 release is not available in the npm registry at the time of this audit. CI still fails on any critical npm advisory, and React Router should be upgraded as soon as a compatible fixed release is published.

## Remaining hardening work

- Replace long-lived JWT storage in `localStorage` with short-lived access tokens and rotated refresh tokens in secure, HttpOnly, SameSite cookies.
- Re-encode uploaded images or scan them before production object storage.
- Replace database error strings returned by legacy handlers with generic client errors and structured server logs.
