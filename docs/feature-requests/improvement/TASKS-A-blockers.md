# Part A - ship blockers (R1-R10)

Statuses live in `BACKLOG.md`. Evidence goes to `LEDGER.md`. Card format: why -> steps -> acceptance -> verify -> human review.

---

## R1 - wire gateway into compose, kill X-User-Id trust

Wave 1 | Effort M | Depends: - | Stephen input: -

Why: `services/gateway/` is built (JWT verification, routing) but `deploy/docker-compose.yml` defines no gateway service. `deploy/README.md:123` admits authsvc, tracksvc, and bff trust the `X-User-Id` header. Any client that can reach a service port can act as any user.

Steps:
1. Read `services/gateway/cmd` and `services/gateway/internal` to confirm route table, JWT audience (`AUTH_AUD` already referenced in compose), and upstream config format.
2. Add a `gateway` service to `deploy/docker-compose.yml`: build from `deploy/Dockerfile.go`, env for upstream URLs + `AUTH_AUD`, healthcheck, restart policy.
3. Move all published host ports off pricesvc/dealsvc/authsvc/tracksvc/bff; keep them on an internal compose network only. Publish only gateway (and web).
4. In gateway, strip any inbound `X-User-Id`/`X-*` identity headers from client requests before setting its own verified values.
5. Update `web` and BFF base URLs to call the gateway. Update `deploy/README.md` §7 and `deploy/.env.example`.
6. Add an integration test: request with forged `X-User-Id` to gateway does not impersonate; direct service access from outside the network fails.

Acceptance: no service except gateway/web/db(optional) publishes a host port; forged identity headers are ignored; `make up && make smoke` green.

Verify: `docker compose -f deploy/docker-compose.yml config` (no stray ports); `curl -H "X-User-Id: 1" http://localhost:<gateway>/v1/track/wishlist` returns 401 without JWT; smoke output in ledger.

Human review: read the compose diff; try the forged-header curl yourself; confirm README §7 no longer lists the gap.

---

## R2 - purge node_modules from git

Wave 1 | Effort S | Depends: - | Stephen input: -

Why: 13,787 of 14,734 tracked files (94%) are `extension/node_modules/` and `web/node_modules/`. Clones and diffs are bloated; secret scanning is noisy; open-sourcing the extension (R36) is impossible like this.

Steps:
1. `git rm -r --cached extension/node_modules web/node_modules`.
2. Add `node_modules/` to root `.gitignore` (verify `services/ml/.venv` pattern is present too).
3. Confirm CI (`.github/workflows/ci.yml`) runs `npm install`/`npm ci` and still passes.
4. Note in ledger: history still contains the blobs; full history rewrite is a separate Stephen decision (record as optional follow-up, do not rewrite history unilaterally).

Acceptance: `git ls-files | grep -c node_modules/` returns 0; CI green.

Verify: the grep above plus CI run link/output in ledger.

Human review: confirm working tree still builds locally (`make test-web`), decide whether to schedule a history rewrite (BFG) before R36.

---

## R3 - add LICENSE (proprietary core + OSS extension)

Wave 1 | Effort S | Depends: - | Stephen input: decision (license choice)

Why: no LICENSE exists anywhere. The PRD's trust moat claims an open-source extension; without a license that claim is legally empty. Default all-rights-reserved also blocks store reviewers and community contributions.

Steps:
1. Draft root `LICENSE` (proprietary, CyberSkill Software Solutions Consultancy and Development JSC) covering services/, web/, db/, deploy/.
2. Draft `extension/LICENSE` as MIT (recommended; Apache-2.0 acceptable) with CyberSkill copyright line.
3. Add root `NOTICE.md` explaining the split and third-party dependency licenses (generate a summary with `license-checker` for JS and `go-licenses` for Go).
4. Mark task `needs_stephen` with the concrete ask: approve MIT for extension, proprietary for the rest.

Acceptance: both files present, NOTICE lists the split, Stephen approval recorded in ledger.

Verify: files exist; dependency license report attached in ledger.

Human review: Stephen approves or changes the license choice; check the copyright entity name is the legal name.

---

## R4 - real root metadata, kill "Create Next App"

Wave 1 | Effort S | Depends: - | Stephen input: -

Why: `web/app/layout.tsx:17-18` still ships `title: "Create Next App"`. Every SERP snippet, browser tab, and shared link shows scaffold text.

Steps:
1. Replace metadata in `web/app/layout.tsx`: title template (`%s | SănDeal`), VN default description (one sentence: theo dõi giá + phát hiện sale ảo trên Shopee, TikTok Shop, Lazada), `metadataBase` from env, canonical, OG image (create a simple 1200x630 static asset), Twitter card, `lang="vi"` on html.
2. Ensure `(marketing)` pages keep their own metadata overrides (they already set some).
3. Add a jest/snapshot test asserting root metadata does not contain "Create Next App".

Acceptance: rendered `<head>` shows real brand metadata on `/login` and a marketing page.

Verify: `npm test` in web/; `curl -s localhost:3000/login | grep -i "<title>"` output in ledger.

Human review: check the Vietnamese copy reads naturally; confirm OG image renders in a link-preview checker.

---

## R5 - auth guard for the whole (app) route group

Wave 1 | Effort S | Depends: - | Stephen input: -

Why: `web/middleware.ts` only guards paths starting with `/dashboard`. `/wishlist`, `/alerts`, `/products/[id]/chart` render without any session cookie.

Steps:
1. Extend the middleware check to a protected-prefix list: `/dashboard`, `/wishlist`, `/alerts`, `/products`.
2. Keep the `next=` redirect param behavior (DEC-WEB-05, HTTP 307) for all of them.
3. Add tests: unauthenticated request to each protected path redirects to `/login?next=...`; public marketing paths pass through.

Acceptance: all four prefixes redirect when `refresh_token` cookie is absent.

Verify: web jest suite green; curl each path without cookies, note 307 + Location in ledger.

Human review: click through logged-out in a browser; confirm no protected page flashes content before redirect.

---

## R6 - CSRF + origin checks + login rate limit

Wave 1 | Effort M | Depends: R1 | Stephen input: -

Why: the login POST (`web/app/(auth)/login/page.tsx:20-23`) and the refresh route (`web/app/api/auth/refresh/route.ts`) carry no CSRF protection; nothing rate-limits `/v1/auth/login`.

Steps:
1. On the BFF/API routes that set or use the refresh cookie: enforce `Origin`/`Sec-Fetch-Site` checks; reject cross-site requests.
2. Set the refresh cookie `SameSite=Strict` (verify current attributes; fix if weaker), `Secure`, `HttpOnly`, scoped path.
3. Add a double-submit CSRF token (or keep pure origin-check strategy if documented) for state-changing BFF routes; document the chosen strategy in `docs/conventions/`.
4. At the gateway (R1), add per-IP token bucket on `/v1/auth/login` and `/v1/auth/refresh` (e.g., 10/min with burst), returning 429 with Retry-After; reuse the existing rate-limit middleware found in the gateway if present.
5. Tests: cross-origin refresh rejected; 11th login attempt in a minute gets 429.

Acceptance: cross-site refresh fails; login brute force throttled; legit flow unaffected (`make smoke` + manual login).

Verify: curl matrix (same-origin ok, cross-origin 403, 11th attempt 429) recorded in ledger.

Human review: log in once through the browser; review cookie attributes in devtools; check the 429 message is friendly VN text at the web layer.

---

## R7 - session restore on page reload

Wave 1 | Effort S | Depends: - | Stephen input: -

Why: the access token lives in a module variable (`web/lib/auth.ts:17-19`). A page reload drops it even though the httpOnly refresh cookie is still valid, so every reload behaves like a logout.

Steps:
1. Add a bootstrap step (app-shell effect or server component pattern): if no in-memory access token and a refresh cookie exists, call the refresh endpoint once, store the new access token in memory.
2. Handle refresh failure by clearing state and redirecting to `/login?next=`.
3. Prevent thundering refresh: single-flight the bootstrap call.
4. Tests: reload with valid cookie restores an authenticated fetch; reload with expired cookie lands on login.

Acceptance: reloading `/wishlist` while logged in stays logged in without user action.

Verify: web jest green; manual reload demo noted in ledger (screenshot or curl trace).

Human review: reload each app page logged-in; confirm no visible logout flicker.

---

## R8 - extension env-config endpoints + domain decision

Wave 1 | Effort S | Depends: - | Stephen input: decision (final domain/brand)

Why: `extension/src/sync/ws-client.ts:1` hardcodes `wss://api.sandeal.vn/v1/ext/ws`. Staging/dev testing is impossible, and the domain itself (sandeal.vn) has not been confirmed as owned. Store listings will freeze the name (R58 depends on this).

Steps:
1. Introduce a build-time config module (`extension/src/shared/config.ts`) with dev/staging/prod endpoint sets selected by build env; default prod.
2. Replace all hardcoded URLs (grep the whole `extension/src` for `sandeal.vn` and `http`).
3. Keep DNR allowlists in sync with the config (rules.json per env or documented single-domain policy).
4. Record the ask for Stephen: confirm ownership of sandeal.vn (or pick final domain) before R49/R58.

Acceptance: `npm run build` (or the repo's build script) can produce dev and prod bundles pointing at different endpoints; tests green.

Verify: grep output showing no hardcoded prod URL outside config; both bundles' effective URLs in ledger.

Human review: Stephen confirms the domain; reviewer checks DNR rules still block non-allowlisted hosts (ties into R31).

---

## R9 - guard demo seed against prod

Wave 1 | Effort S | Depends: - | Stephen input: -

Why: `make seed` inserts demo user 999 / product 100 into whatever DB compose points at. One wrong shell on a prod host pollutes real data.

Steps:
1. Add an environment gate to the seed target in `Makefile`: refuse unless `APP_ENV=dev` (or `ALLOW_SEED=1`) is set; print a clear refusal otherwise.
2. Mirror the same guard inside `make smoke` (it calls seed).
3. Tag the demo rows (e.g., negative IDs or a `demo` flag column if trivial) so cleanup is possible; at minimum document their IDs in `deploy/README.md`.

Acceptance: `make seed` on a shell without the env var refuses; with it, works as before.

Verify: both invocations' output in ledger; `make smoke` still green with the var set.

Human review: run `make seed` without the var on your Mac; confirm refusal message is unmistakable.

---

## R10 - Go toolchain alignment + govulncheck in CI

Wave 1 | Effort S | Depends: - | Stephen input: -

Why: code builds on Go 1.25 while `docs/feature-requests/SHIP-GUIDE.md` mandates 1.22 (`docs/AUDIT-REPORT.md` hygiene section). Drift between the contract and reality confuses every future agent. No vulnerability scanning runs today.

Steps:
1. Update SHIP-GUIDE to declare Go 1.25 (recommended - code already there and CI uses 1.25) OR pin everything to 1.22; pick one, log rationale.
2. Align every `go.mod` `go` directive and the CI setup-go version to the chosen version.
3. Add a `govulncheck ./...` step per Go module in CI (allow failure = false).
4. Fix or document any findings.

Acceptance: SHIP-GUIDE, go.mod files, and CI agree on one version; govulncheck green (or findings triaged in ledger).

Verify: grep of `go 1.` across go.mod files; CI run output.

Human review: skim govulncheck findings triage; approve the version decision line in SHIP-GUIDE.
