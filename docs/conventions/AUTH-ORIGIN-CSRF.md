# Auth origin / CSRF strategy (R6)

## Choice

**Pure origin-check** on Next.js route handlers that set or clear the refresh cookie. No double-submit CSRF token.

Rationale: the refresh credential is `HttpOnly` + `SameSite=Strict` (and `__Host-` prefixed in production), so a cross-site form cannot read or send it as a classic cookie CSRF. Browser `fetch` from another origin either omits cookies under Strict or is rejected by our Origin check before the handler touches auth upstream.

## Where enforced

| Route | Check |
|-------|--------|
| `POST /api/auth/login` | `requestHasAllowedOrigin` → 403 |
| `POST /api/auth/register` | same |
| `POST /api/auth/refresh` | same |
| `POST /api/auth/logout` | same |

Implementation: [`web/lib/server-auth.ts`](../../web/lib/server-auth.ts) `requestHasAllowedOrigin`.

- Production: `Origin` must equal `APP_ORIGIN` when configured, else the request’s own origin.
- Missing `Origin`: allowed only when `NODE_ENV !== "production"` (local tooling).

## Cookie attributes

See `refreshCookieOptions` in `server-auth.ts`:

- `httpOnly: true`
- `sameSite: "strict"`
- `secure: true` in production
- Production name: `__Host-sandeal_refresh` (host-only, no Domain)

## Gateway rate limits (per-IP, 60s window)

| Path | Limit |
|------|------:|
| `POST /v1/auth/login` | 5 |
| `POST /v1/auth/refresh` | 10 |
| other routes | 100 (authenticated: per-user) |

Limiter fails closed on Redis errors for these paths (503). See [`services/gateway/internal/gw/ratelimit.go`](../../services/gateway/internal/gw/ratelimit.go).

## Explicitly out of scope here

- Double-submit CSRF tokens (revisit if we ever set cookies from cross-site OAuth redirects without SameSite=Strict).
- Adaptive / risk-based throttling (TASK-TRUST-004).
