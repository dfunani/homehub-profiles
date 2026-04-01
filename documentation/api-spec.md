# HomeHub Profiles — API Specification (Outline)

**Version:** 0.1 (draft)  
**Base URL:** `https://api.example.com` (configure per environment)  
**Format:** JSON UTF-8  
**Auth:** `Authorization: Bearer <access_token>` unless noted

This is an **outline** for implementation; finalize field names, pagination, and error codes in OpenAPI/Swagger when the stack is chosen.

---

## Conventions

### Timestamps

ISO 8601 in UTC, e.g. `2026-03-31T12:00:00Z`.

### IDs

UUIDs in path and JSON.

### Pagination (list endpoints)

Query: `limit` (default 20, max 100), `cursor` (opaque).  
Response: `{ "data": [...], "next_cursor": "..." | null }`.

### Errors

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "title is required",
    "details": {}
  }
}
```

Suggested `code` values: `VALIDATION_ERROR`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `INTERNAL`.

---

## Health

### `GET /health`

No auth. Returns `200` with `{ "status": "ok" }`. Optionally include DB connectivity.

---

## Auth

### `POST /auth/register`

Create account.

**Body:**

```json
{
  "email": "user@example.com",
  "password": "string",
  "role": "client | provider",
  "display_name": "string"
}
```

**201:** `{ "user": { ... }, "access_token": "...", "token_type": "Bearer", "expires_in": 3600 }`

**409:** email already registered.

---

### `POST /auth/login`

**Body:** `{ "email": "...", "password": "..." }`

**200:** same token envelope as register.

**401:** invalid credentials.

---

### `POST /auth/refresh` (optional MVP)

**Body:** `{ "refresh_token": "..." }`

**200:** new access token.

---

## Current user

### `GET /me`

**Auth:** required.

**200:**

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "role": "client | provider",
  "profile": {
    "display_name": "string",
    "contact_email": "string | null",
    "contact_phone": "string | null",
    "bio": "string | null"
  },
  "created_at": "ISO8601"
}
```

---

### `PATCH /me/profile`

**Auth:** required.

**Body (partial):** `display_name`, `contact_email`, `contact_phone`, `bio`

**200:** updated profile object.

---

## Users (public / limited)

### `GET /users/{userId}`

Public profile for providers (and optionally clients). Hide email unless policy allows.

**200:**

```json
{
  "id": "uuid",
  "display_name": "string",
  "role": "provider",
  "bio": "string | null"
}
```

**404:** not found.

---

## Listings

### `GET /listings`

Search/discover. **Auth:** TBD (public or bearer).

**Query:** `q` (search string), `limit`, `cursor`

**200:** paginated `Listing` summaries:

```json
{
  "data": [
    {
      "id": "uuid",
      "title": "string",
      "description": "string",
      "price_text": "string",
      "availability_text": "string",
      "provider": {
        "id": "uuid",
        "display_name": "string"
      },
      "created_at": "ISO8601"
    }
  ],
  "next_cursor": null
}
```

---

### `POST /listings`

**Auth:** required, role `provider`.

**Body:**

```json
{
  "title": "string",
  "description": "string",
  "price_text": "string",
  "availability_text": "string"
}
```

**201:** full listing including `id`, `created_at`.

---

### `GET /listings/{listingId}`

**200:** full listing + provider summary.

**404:** not found.

---

### `PATCH /listings/{listingId}`

**Auth:** provider who owns the listing.

**Body:** partial fields same as create.

**200:** updated listing.

**403 / 404:** as appropriate.

---

### `DELETE /listings/{listingId}`

**Auth:** owning provider.

**204:** no content.

---

### `GET /users/{userId}/listings`

Listings for a provider (public catalog).

**200:** paginated listings.

---

## Conversations

### `POST /conversations`

Start a conversation about a listing.

**Auth:** required, role `client`.

**Body:**

```json
{
  "listing_id": "uuid"
}
```

**201:**

```json
{
  "id": "uuid",
  "listing_id": "uuid",
  "client_user_id": "uuid",
  "provider_user_id": "uuid",
  "created_at": "ISO8601",
  "last_message_at": "ISO8601 | null"
}
```

**409:** conversation already exists for this client + listing.

**403:** client is the provider of the listing.

---

### `GET /conversations`

**Auth:** required. Lists conversations for the current user.

**200:** paginated conversation summaries (include other participant name, listing title, last message preview).

---

### `GET /conversations/{conversationId}`

**Auth:** participant only.

**200:** conversation metadata + listing snippet.

---

## Messages

### `GET /conversations/{conversationId}/messages`

**Auth:** participant.

**Query:** `limit`, `cursor` (chronological; define whether cursor is message id or time).

**200:** paginated messages:

```json
{
  "data": [
    {
      "id": "uuid",
      "sender_user_id": "uuid",
      "body": "string",
      "created_at": "ISO8601"
    }
  ],
  "next_cursor": null
}
```

---

### `POST /conversations/{conversationId}/messages`

**Auth:** participant.

**Body:** `{ "body": "string" }`

**201:** created message.

**400:** empty body / too long.

---

## Webhooks / real-time (out of scope for REST MVP)

Document separately if adding **WebSocket** `GET /ws` or **SSE** for live messages.

---

## Implementation checklist

- [ ] Register/login + JWT middleware
- [ ] Migrations for users, profiles, listings, conversations, messages
- [ ] Listing CRUD + ownership checks
- [ ] Search (simple `ILIKE` or FTS)
- [ ] Conversation create + idempotency/unique constraint
- [ ] Message list + post
- [ ] Rate limiting on auth and message endpoints
- [ ] Export OpenAPI 3.1 from code or hand-written YAML
