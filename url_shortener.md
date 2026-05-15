# URL Shortener with Click Analytics
## LineSpec Template — Product Requirements Document

---

## Overview

A self-hosted URL shortener API with click analytics. Users create short slugs that redirect to destination URLs, and every redirect is tracked for later reporting. Management operations (create, delete, view analytics) require authentication. The redirect path is optimized for speed — analytics recording must never block or delay a redirect response.

This document defines the behavioral contracts the template enforces. It is stack-agnostic. Any implementation that satisfies these contracts is correct.

---

## Entities

**Link**
- `slug` — unique identifier used in the redirect URL
- `destination` — the full URL to redirect to
- `created_at`
- `created_by` — the authenticated user who created it

**Click**
- `slug` — the link that was accessed
- `clicked_at`
- `user_agent` — from request headers, nullable
- `referer` — from request headers, nullable
- `ip_address` — from request headers, nullable

---

## Endpoints

### `GET /:slug` — Redirect

Resolves a slug and returns a redirect to the destination URL. Records a click asynchronously. This endpoint is unauthenticated and must be as fast as possible.

**Behavioral contracts:**
- Returns `301` with a `Location` header set to the destination URL
- Click recording must be fire-and-forget — the response must not wait for the write to complete
- Returns `404` if the slug does not exist
- Returns `404`, not `500`, if click recording fails — the redirect succeeds regardless of analytics state

---

### `POST /links` — Create Link

Creates a new short link. Requires authentication.

**Behavioral contracts:**
- Returns `401` if no valid bearer token is provided, without touching the database
- Returns `201` with the created link object on success
- Returns `409` if the requested slug already exists
- If no slug is provided, generates a random one that does not collide with existing slugs
- The `created_by` field is populated from the authenticated user identity, not from the request body

---

### `DELETE /links/:slug` — Delete Link

Deletes a link. Requires authentication. A user may only delete links they own.

**Behavioral contracts:**
- Returns `401` if no valid bearer token is provided
- Returns `404` if the slug does not exist
- Returns `403` if the slug exists but was created by a different user
- Returns `204` on success
- Associated click records are deleted as part of the same transaction — partial deletion is not a valid outcome

---

### `GET /links/:slug/analytics` — Click Analytics

Returns aggregate and recent click data for a link. Requires authentication. A user may only view analytics for links they own.

**Behavioral contracts:**
- Returns `401` if no valid bearer token is provided
- Returns `404` if the slug does not exist
- Returns `403` if the slug exists but was created by a different user
- Returns `200` with a response body containing:
  - `total_clicks` — integer count of all recorded clicks
  - `clicks_last_24h` — integer count of clicks in the last 24 hours
  - `recent_clicks` — array of the 10 most recent click records, ordered by `clicked_at` descending

---

### `GET /links` — List Links

Returns all links created by the authenticated user.

**Behavioral contracts:**
- Returns `401` if no valid bearer token is provided
- Returns `200` with an array of link objects belonging to the authenticated user
- Links created by other users are never included in the response

---

## Authentication

Authentication is handled by a separate identity service. The API validates bearer tokens by calling the identity service synchronously before processing any authenticated request. The identity service is an external dependency — it is not implemented as part of this template.

**Behavioral contracts:**
- Token validation must occur before any database read or write on authenticated endpoints
- A `503` response must be returned if the identity service is unreachable — silent failure or default-authenticated behavior is not acceptable
- A `401` response must be returned if the identity service returns an invalid or expired token response

---

## Provenance Record Hierarchy

### Brief — Self-Hosted Link Management

> Engineers and developers need a simple, self-hosted alternative to commercial link shorteners. The system must be fast on the redirect path and must not compromise redirect reliability for the sake of analytics.

### Blueprint — Redirect Behavior

> Slugs resolve to destination URLs and return a permanent redirect. Click recording is decoupled from the redirect response path. The redirect must succeed even if analytics infrastructure is degraded.

**Constraints:**
- MUST return 301 with a correct Location header for known slugs
- MUST return 404 for unknown slugs
- MUST NOT block the redirect response on click write completion
- MUST record a click attempt regardless of whether the write succeeds
- MUST return 404, not 500, if the click write fails

### Blueprint — Link Lifecycle Management

> Authenticated users can create, list, and delete their own links. Ownership is enforced at the API layer — users cannot view or modify links they do not own.

**Constraints:**
- MUST validate bearer tokens against the identity service before any database operation
- MUST return 401 before touching the database when no valid token is present
- MUST return 503 if the identity service is unreachable
- MUST populate created_by from the verified token identity, not from request input
- MUST return 409 on slug collision
- MUST return 403 when a user attempts to delete or view analytics for a link they do not own
- MUST delete associated click records in the same transaction as the link — partial deletion MUST NOT occur

### Blueprint — Click Analytics

> Link owners can retrieve aggregate and recent click data for their links. Analytics data is derived from the clicks table and scoped to the requesting user's links.

**Constraints:**
- MUST enforce ownership before returning analytics data
- MUST return total click count, clicks in the last 24 hours, and the 10 most recent click records
- MUST order recent clicks by clicked_at descending

### Blueprint — Identity Service Dependency

> Token validation is delegated entirely to an external identity service. This service is the sole authority on token validity. The API must not implement fallback authentication or cache validation results.

**Constraints:**
- MUST call the identity service synchronously on every authenticated request
- MUST NOT cache token validation results between requests
- MUST return 503, not a default authenticated state, when the identity service is unreachable

---

## Behavioral Specifications

The following scenarios must be covered by associated specs. Each maps to one or more blueprint constraints above.

| Scenario | Key Assertions |
|---|---|
| Redirect — known slug | Returns 301, Location header matches destination |
| Redirect — unknown slug | Returns 404 |
| Redirect — click not in critical path | Response does not wait on click write; click recorded asynchronously |
| Redirect — click write failure | Returns 301 regardless; no 500 |
| Create link — authenticated | Returns 201 with link object; created_by from token |
| Create link — unauthenticated | Returns 401; no database write occurs |
| Create link — identity service down | Returns 503; no database write occurs |
| Create link — slug collision | Returns 409 |
| Create link — no slug provided | Returns 201; slug is generated and unique |
| Delete link — owner | Returns 204; link and clicks deleted in same transaction |
| Delete link — non-owner | Returns 403 |
| Delete link — unknown slug | Returns 404 |
| Delete link — unauthenticated | Returns 401 |
| Analytics — owner | Returns 200 with correct total, 24h count, and recent clicks |
| Analytics — non-owner | Returns 403 |
| Analytics — unauthenticated | Returns 401 |
| List links — authenticated | Returns only the requesting user's links |
| List links — unauthenticated | Returns 401 |

---

## Out of Scope

The following are intentionally excluded to keep the template bounded and the agent's task completable in a single session:

- Custom domain support
- Link expiry
- Password-protected links
- QR code generation
- Rate limiting on the redirect endpoint
- The identity service implementation — token validation is mocked in specs
