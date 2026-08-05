# VeloHam Backend Architecture

VeloHam is structured as a modular monolith. It runs as one Go process today, but each business area has its own module boundary so it can later move to a separate service with minimal code movement.

## Runtime

- One Go API process
- PostgreSQL as source of truth
- Redis-ready cache/notification abstraction in `internal/common/cache`
- WebSocket endpoint for chat
- JWT auth
- Versioned API under `/api/v1`
- Legacy `/api` routes kept for compatibility with the current frontend

## Module Boundary

New code should live under:

```text
internal/modules/<module>/
```

Each module should evolve toward this shape:

```text
handlers.go      HTTP transport only
routes.go        route registration only
service.go       business rules
repository.go    database access
models.go        module DTOs/domain models when not shared
```

Current modules:

- `auth`
- `users`
- `listings`
- `categories`
- `parts`
- `chat`
- `messages`
- `search`
- `buy_requests`
- `favorites`
- `notifications`
- `reviews`
- `admin`
- `uploads`

## Common Package

Shared infrastructure belongs in `internal/common`:

- `middleware`
- `logger`
- `config`
- `utils`
- `validator`
- `apperrors`
- `constants`
- `response`
- `cache`

Do not put business rules in `common`.

## Dependency Rule

Allowed:

- module -> `common`
- module -> shared persistence models while the monolith is being migrated
- routes composition root -> all modules

Avoid:

- module -> another module's handler
- business logic in route registration
- direct DB access from React-facing route glue once a repository exists

## Microservice Extraction Path

Good first extraction candidates:

1. `chat`: WebSocket, message history, unread counts
2. `search`: fast filtering, Redis cache, later full-text/index service
3. `notifications`: events, unread notification inbox
4. `auth`: token issuing, refresh token rotation, email confirmation

Extraction plan:

1. Keep the same REST contract under `/api/v1/<module>`.
2. Move `repository/service/handler/routes` package into a new service.
3. Replace direct local calls with HTTP/gRPC/event calls only at the boundary.
4. Keep PostgreSQL ownership clear before splitting databases.

## API Versioning

New clients should use `/api/v1/...`.

Existing `/api/...` endpoints remain active until the frontend is migrated.
