# HomeHub Profiles — Project

## Purpose

**HomeHub Profiles** (module `dfunani/homehub-profiles`) is a self-contained backend application for a **home-services marketplace**: people who offer household work (cleaning, gardening, laundry, handyman tasks, and similar) can **register as service providers**, publish **service listings**, and **communicate with clients** who are looking for those services.

The server exposes a clear HTTP API. It is **not** designed around any specific frontend (for example, the separate `homehub-web` prototype). Any client—web, mobile, or another service—can integrate using the same contracts.

## Relationship to `homehub-web`

The web app in `../homehub-web` is a **reference for product shape** only: it models *buyers* vs *sellers*, searchable *services*, and *chats* tied to a service. This backend **does not** assume that UI stack, routing, or mock `localStorage` behavior. It implements the **real** persistence, authentication, and API boundaries those flows imply.

## What the service is (conceptual)

1. **Users** sign up and authenticate. Each user has a **role** (at minimum: **client** seeks services; **provider** offers services). A user might be extended later to support both roles; the initial scope can keep a single role per account for simplicity.

2. **Service providers** maintain a **profile** (display name, optional contact fields) and create **listings**: title, description, price (free-text or structured later), and availability notes.

3. **Clients** **discover** listings (search/filter), open a **listing detail**, and can **start or continue a conversation** with the provider about that listing. Messaging is **per conversation**, linked to a listing and the two parties.

4. The system **does not** have to implement payments, scheduling, or job completion in v1; those are natural later layers. v1 focuses on **identity**, **profiles**, **listings**, and **messaging**.

## Goals

- **API-first**: stable resources and error shapes suitable for multiple clients.
- **Self-contained**: runnable server with its own configuration, database, and deployment story.
- **Clear boundaries**: auth, user/profile, listings, and chat are separable modules for testing and evolution.

## Non-goals (initial release)

- No hard dependency on a specific SPA framework or Next.js.
- No requirement to mirror the mock’s anonymous-auth behavior; real auth (e.g. email/password or OAuth) is specified in the design and API docs.
- Payment processing, calendar booking, and ratings are **out of scope** unless explicitly added to a later milestone.

## Documentation index

| Document | Contents |
|----------|----------|
| [project.md](./project.md) | This file — vision, scope, terminology. |
| [design-doc.md](./design-doc.md) | Architecture, data model, security, operations. |
| [api-spec.md](./api-spec.md) | HTTP API outline (resources, methods, payloads). |
| [migrations.md](./migrations.md) | SQL migrations (golang-migrate), upgrade/downgrade commands. |

## Terminology

| Term | Meaning |
|------|---------|
| **Client** | User looking to hire someone for a home service (analogous to *buyer* in the web mock). |
| **Provider** | User who offers services (analogous to *seller* in the mock). |
| **Listing** | A provider’s offered service (name, description, price, availability). |
| **Conversation** | Chat thread between a client and a provider, associated with one listing. |

## Success criteria (MVP)

- Providers can register, authenticate, manage profile, and CRUD listings.
- Clients can register, authenticate, search listings, and view listing + provider summary.
- Either party can participate in listing-scoped conversations and exchange messages.
- Documentation is sufficient for another developer to implement the server without reading the web repo’s UI code.
