# HomeHub Profiles — Design Document

## 1. Overview

This document describes how **HomeHub Profiles** should be structured: components, data model, security, and cross-cutting concerns. It complements [project.md](./project.md) and [api-spec.md](./api-spec.md).

## 2. Architectural style

- **Monolith first**: single deployable process (e.g. Go `main`) with internal packages for `auth`, `users`, `listings`, `conversations`, `messages`.
- **REST over HTTPS**: JSON request/response bodies; resource-oriented URLs (see API spec).
- **Stateless API layer**: each request carries credentials (e.g. `Authorization: Bearer <JWT>`); server validates token and loads actor context.
- **Database-backed**: PostgreSQL (recommended) or SQLite for local dev; no in-memory-only production store.

Optional later: extract notification or search to workers; not required for MVP.

## 3. High-level components

```
Clients (any HTTP)
        │
        ▼
┌───────────────────┐
│  TLS / HTTP       │
│  Router + MW      │  ← request ID, logging, CORS, auth
└─────────┬─────────┘
          │
┌─────────▼─────────┐
│  Handlers         │  ← validate input, call services
└─────────┬─────────┘
          │
┌─────────▼─────────┐
│  Domain services  │  ← business rules (who can delete listing, etc.)
└─────────┬─────────┘
          │
┌─────────▼─────────┐
│  Repository layer │  ← SQL / transactions
└─────────┬─────────┘
          │
┌─────────▼─────────┐
│  PostgreSQL       │
└───────────────────┘
```

## 4. Data model (logical)

Entities and relationships (names are indicative; exact columns belong in migrations).

### 4.1 `users`

- `id` (UUID, PK)
- `email` (unique, nullable if phone-auth added later)
- `password_hash` (if using password auth) or linkage to OAuth subject
- `role` — enum: `client` | `provider` (extend later if needed)
- `created_at`, `updated_at`

### 4.2 `profiles`

One row per user (or merge into `users` if preferred).

- `user_id` (FK → users)
- `display_name`
- `contact_phone` / `contact_email` (optional, for provider visibility rules)
- `bio` (optional)

### 4.3 `listings`

- `id` (UUID)
- `provider_user_id` (FK → users; must be role `provider` or enforce at service layer)
- `title`, `description`
- `price_text` (MVP) — free text, e.g. `R250/hour`; structured money later
- `availability_text`
- `created_at`, `updated_at`
- `deleted_at` (soft delete, optional)

Index: `provider_user_id`, full-text or `ILIKE` on `title`/`description` for search (upgrade to proper search engine later).

### 4.4 `conversations`

Represents one thread per **(listing, client, provider)** triple (or listing + two participants).

- `id` (UUID)
- `listing_id` (FK)
- `client_user_id`, `provider_user_id` (FKs)
- `created_at`
- `last_message_at` (denormalized for sorting)
- Unique constraint on `(listing_id, client_user_id)` so one conversation per client per listing.

### 4.5 `messages`

- `id` (UUID)
- `conversation_id` (FK)
- `sender_user_id` (FK)
- `body` (text; cap length e.g. 8–16k chars)
- `created_at`

Index: `(conversation_id, created_at)`.

## 5. Authentication and authorization

### 5.1 Authentication

- **MVP**: email + password (bcrypt/argon2), issue **JWT** (access token, short TTL) and optionally refresh token stored server-side or in rotating cookies.
- **Alternative**: magic link or OAuth2 — same JWT issuance after identity proof.

### 5.2 Authorization rules (examples)

| Action | Rule |
|--------|------|
| Create listing | Authenticated user with role `provider`. |
| Update/delete listing | Only `provider_user_id` of that listing. |
| Search listings | Authenticated or public (product decision); document in API spec. |
| Start conversation | Authenticated `client`; listing exists; client ≠ provider; no duplicate per listing. |
| Send message | User is participant of conversation. |
| Read messages | User is participant. |

## 6. Validation and errors

- Centralized validation for required fields, string lengths, UUID format.
- HTTP status: `400` validation, `401` unauthenticated, `403` forbidden, `404` missing resource, `409` conflict (e.g. duplicate conversation), `500` unexpected.
- JSON error body: `{ "error": { "code": "string", "message": "human readable" } }` (exact shape fixed in API spec).

## 7. Observability

- Structured logging (request ID, method, path, status, duration).
- Health endpoint: `GET /health` → DB ping optional.
- Metrics (optional MVP): request counts, latency histograms.

## 8. Configuration

- Environment variables: `DATABASE_URL`, `JWT_SECRET` (or key path), `HTTP_ADDR`, `CORS_ALLOWED_ORIGINS` (if needed).
- No secrets in repo; use `.env` locally and secret manager in production.

## 9. Testing strategy

- **Unit tests**: domain rules (e.g. “provider cannot message self”).
- **Integration tests**: HTTP against real DB (testcontainer or embedded SQLite) for happy paths and auth failures.

## 10. Deployment assumptions

- Single binary + migrations (e.g. `golang-migrate` or embedded SQL).
- Horizontal scaling: stateless API + shared DB; sticky sessions not required if using JWT.

## 11. Future extensions (not MVP)

- Payments, escrow, invoices
- Booking / calendar with time zones
- Reviews and ratings
- Push notifications / email digests
- Admin moderation and reports
- Provider verification (ID, background check)

## 12. Open decisions

Record choices as you implement:

- Public vs authenticated search
- Whether one user can be both client and provider (schema + role model)
- Message real-time delivery (polling vs WebSocket vs SSE)
