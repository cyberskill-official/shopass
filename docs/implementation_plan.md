# Comprehensive Investigation & Improvement Plan

This document outlines the findings from a deep, comprehensive investigation into the backend services. While the 90 FRs were marked as "done" because the integration tests and builds succeeded, several critical modules are currently relying on **mocked data, dummy implementations, or incomplete TODOs** and are not fully production-ready.

## User Review Required
> [!IMPORTANT]
> Please review this plan carefully. Once approved, I will implement the real logic to replace the mocks across all services and groom the backlog to reflect the actual remaining work.

## Open Questions
> [!WARNING]
> 1. **Scraping Farm (services/scrape)**: Should I implement the TikTok scraping logic using a local Go-Playwright script, or is there an external Node.js farm I should call via HTTP/gRPC?
> 2. **Proxy Provider (services/scrape)**: For `proxy/pool.go`, do we have a specific provider (like BrightData/Oxylabs) to integrate with, or should I load a real list of proxies from the database?
> 3. **Notification System (services/deal)**: Should `dealsvc` push notifications directly to a Kafka/Redis queue, or make an HTTP call to `services/notif`?
> 4. **Comply Database (services/comply)**: Should I create the necessary PostgreSQL tables for `ecom_obligations` and `tx_counts`, or are they already defined somewhere?

## Proposed Changes

### services/deal (FR-DEAL-001, FR-DEAL-006)
- **Replace Mocks**: Modify `internal/deal/service.go` to query real price history and product data from the database (or via `scrape` service) instead of using `fakeProduct` and hardcoded price arrays.
- **Implement Real Notifications**: In `cmd/dealsvc/main.go`, replace `dummyNotif` with a real client that publishes messages to `services/notif` for nightly scoring alerts.

### services/bill (FR-BILL-001, FR-BILL-003)
- **Activate Subscriptions**: Update `internal/bill/reconcile.go` and `internal/api/ipn.go` to remove the `// TODO: Activate Subscription` comments. Implement the actual database logic to upgrade a user's `tier` upon successful payment verification.

### services/scrape (FR-SCRAPE-007, FR-SCRAPE-008)
- **Real TikTok Scraping**: Update `internal/adapters/tiktok/adapter.go` to replace the hardcoded `PriceSnapshot{Price: 99000}` with a real scraping invocation.
- **Real Proxy Rotation**: Update `internal/proxy/pool.go` to integrate with a real proxy provider or fetch real IPs from the DB instead of generating dummy `http://proxy.tier.country` strings.

### services/cart (FR-CART-002)
- **Enforce JWT Auth**: In `internal/api/snapshot.go`, remove the dummy `userIDVal = int64(1)` bypass. Wire up the actual JWT authentication middleware to extract and validate the `user_id` from the request context.

### services/comply (FR-COMPLY-002)
- **Implement Database Repository**: Refactor `internal/ecom/repo.go` and `threshold.go` to connect to PostgreSQL via `pgxpool`. Replace the mocked in-memory maps (`txCounts`, `thresholds`, `obs`) with actual SQL queries.

### Documentation & Backlog Grooming
- **Revert FR Statuses**: I will update `BACKLOG.md` and `IMPLEMENTATION-ORDER.md` to move the partially implemented FRs (where we relied on mocks) from `done` back to `in_progress`.
- **Cleanup**: Remove any orphaned test files or leftovers that were scaffolding the mocked implementations.

## Verification Plan

### Automated Tests
- Remove all "dummy" test structs and mock bypasses.
- Write new integration tests that use real DB connections (`go-sqlmock` or Testcontainers) and real JWTs.
- Run `go test ./...` in each service to ensure the real logic passes.

### Manual Verification
- Deploy the backend locally and manually trigger an IPN callback to verify the subscription tier is updated in the DB.
- Use `curl` to call the cart API without a JWT and verify it returns `401 Unauthorized`.
