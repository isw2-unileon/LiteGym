# ADR-003: Cookie sessions with a custom HMAC-signed token

## Status

Accepted

## Date

26-03-2026

## Context

The application needs authenticated sessions for a browser frontend talking to a Go
API. We did not want to depend on an external identity provider for a self-contained
project, and we wanted to avoid storing tokens where browser JavaScript could read
them.

## Decision

Use backend-managed login with a custom token that is structured like a JWT (header,
claims, HMAC SHA-256 signature) but implemented directly in the backend instead of
pulling in a JWT library. The token is delivered to the browser in an `HttpOnly`
cookie.

- Token creation and validation live in
  `backend/internal/service/token_service.go`.
- The cookie is set with `HttpOnly = true`, `SameSite = Lax`, `Path = /`, and
  `Secure`/`MaxAge` driven by configuration.
- Claims carry `sub`, `email`, `username`, `role`, `iss`, `iat`, and `exp`.
- `AuthMiddleware.RequireAuth()` verifies the signature and expiration, and also
  checks that the referenced user still exists so a cryptographically valid token for
  a deleted user is rejected with `401`.
- Authorization uses a lightweight `user` / `admin` role taken from the claims.

Cookie name, secure flag, and token TTL are configurable through
`AUTH_COOKIE_NAME`, `AUTH_COOKIE_SECURE`, and `AUTH_TOKEN_TTL`.

## Consequences

- No external auth dependency and no third-party JWT library to track.
- Browser JavaScript cannot read the token, reducing token-theft surface; fetches to
  protected endpoints must use `credentials: "include"`.
- We own the crypto details and must keep the signing secret safe and the
  verification logic correct.
- The stale-token check adds a database lookup per authenticated request, accepted as
  a correctness-over-latency trade-off.
