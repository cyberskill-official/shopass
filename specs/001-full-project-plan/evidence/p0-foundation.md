# P0 Foundation — Evidence Bundle

**Commit / Build Identifier:** Initial Layer 0 implementation.

**FR IDs Included:**
- FR-EXT-001 (MV3 scaffold) — done
- FR-INFRA-001 (API Gateway/BFF) — done
- FR-INFRA-002 (Data-model foundation) — done
- FR-INFRA-003 (Secrets management) — done

**Test Command Summary:**
- `cd extension && npm install && npm test` — 3 suites, 4 tests, all passed
- `cd extension && npx tsc --noEmit` — no type errors
- `cd services/gateway && GOCACHE=/tmp/go-build-shopass go test -count=1 ./internal/gw/` — 12 handler/middleware tests, all passed
- `cd secrets && GOCACHE=/tmp/go-build-shopass go test -count=1 -v ./...` — 9 provider/cache/mask tests, all passed
- `cd db && GOCACHE=/tmp/go-build-shopass GOMODCACHE=/tmp/go-mod-shopass DATABASE_URL='postgres://shopass:shopass@127.0.0.1:55432/shopass_test?sslmode=disable' go test -count=1 -v ./...` — 8 PostgreSQL-backed migration tests, all passed

**Migration Version:** 3 (0001_extensions, 0002_platform, 0003_app_user_core)

**Dashboard / Log Links:** None yet.

**Security / Compliance Notes:**
- No cleartext secrets: secrets package mask + Vault KV v2 / AWS SM provider interfaces + cache rotation tests (FR-INFRA-003)
- Platform tokens never leave client: extension only sends minimal payload via typed messaging (FR-EXT-001)
- Extension MV3 compliant: ephemeral SW, chrome.storage, alarms >=30s, per-domain host_permissions (FR-EXT-001)
- JWT verify at gateway edge, not self-signed (FR-INFRA-001)
- All money columns will be BIGINT VND (schema conventions documented)
- PDPL compliance: consent framework and no-cleartext enforcement deferred to downstream FRs

**Known Risks and Deferred FRs:**
- PostgreSQL verification used temporary Docker container `shopass-pg-test`, removed after test run
- FR-INFRA-001 Redis integration (rate-limit) uses mock — real Redis needed for production
- FR-INFRA-003 uses HTTP-compatible Vault KV v2 / AWS SM provider code with contract tests; production auth/policy wiring remains environment work
- Layer 0 FRs all marked `done` — next: Layer 1 MVP implementation
