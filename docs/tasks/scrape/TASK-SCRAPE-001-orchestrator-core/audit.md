---
fr_id: TASK-SCRAPE-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Kiểm tra lại từ file hiện tại, không tin audit cũ. TASK-SCRAPE-001 đặc tả orchestrator lõi ở mức triển khai được: 12 mệnh đề §1 normative (đủ MUST, có một SHOULD metric), mỗi mệnh đề testable có AC ở §4 và test ở §5. Schema `scrape_job` khớp DATA-MODEL (PK `product_id`, tier ENUM `scrape_tier` DEFAULT 'cold', `locked_until` cho lock, FK `tracked_product`/`platform`). Tiering hot/warm/cold là hàm thuần `NextRunAt`/`ReTier`; hàng đợi bền re-claim qua `locked_until`; concurrency cap per-platform; ghi qua `InsertSnapshot` của TASK-PRICE-002 chứ không tự viết SQL. Score = 10/10.

## §2 - Findings

### ISS-001 - Arrow glyph trong comment code (đã sửa)
Lượt audit này phát hiện 4 ký tự mũi tên U+2192 trong comment Vietnamese của khối code (`tier.go::ReTier` dòng 175, test dòng 236-237, config dòng 322), vi phạm tiêu chí typography O (chỉ cho phép ASCII `->`). Đã sửa: thay toàn bộ ký tự mũi tên đó thành chuỗi ASCII `->`. Sau khi sửa, quét lại 0 ký tự mũi tên còn lại. Không có em-dash, en-dash, dấu nháy cong, hay ký tự ellipsis ở bất kỳ đâu.

### ISS-002 - Đối chiếu schema scrape_job với DATA-MODEL (xác nhận khớp)
Tự kiểm cột-theo-cột: §3 migration `0001_scrape_job.sql` có `product_id BIGINT PK REFERENCES tracked_product(id)`, `platform_id SMALLINT REFERENCES platform(id)`, `tier scrape_tier NOT NULL DEFAULT 'cold'`, `next_run_at`, `attempts`, `last_status`, `locked_until`. Trùng khít định nghĩa DATA-MODEL của `scrape_job`. Không có khiếm khuyết.

### ISS-003 - Ranh giới ghi DB (xác nhận đúng cross-ref)
§1 #8 + #11 + `pool.go::runOne` gọi `p.price.InsertSnapshot(ctx, snap)` (delta-only TASK-PRICE-002), không có câu `INSERT INTO price_snapshot` nào trong orchestrator; idempotent re-claim dựa `ON CONFLICT DO NOTHING` của PRICE. Cross-ref đúng, không nhân bản logic delta.

## §3 - Traceability §1 -> AC -> artefact

| §1 mệnh đề | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 schema scrape_job + index | AC1 | `0001_scrape_job.sql` |
| #2 tiering tần suất NextRunAt | AC2 | `tier.go::NextRunAt` + `TestNextRunAt_Tiers` |
| #3 re-tier promote/demote | AC3,AC4 | `tier.go::ReTier` + `TestReTier_*` |
| #4 queue bền + re-claim | AC6,AC7 | `queue.go` + `TestPool_ReclaimOrphanJob` |
| #5 retry + backoff giới hạn | AC9 | `pool.go::scheduleRetry` + `TestPool_RetryThenFail` |
| #6 concurrency cap per-platform | AC8 | semaphore + `TestPool_ConcurrencyCapPerPlatform` |
| #7 interface PlatformAdapter | AC10 | `adapter.go` |
| #8 gọi InsertSnapshot (không tự ghi) | AC10 | `pool.go::runOne` |
| #9 cập nhật next_run_at | AC2 | `commit` + NextRunAt |
| #10 OTel metric (SHOULD) | AC12 | counters/gauge |
| #11 idempotent re-claim | AC11 | ON CONFLICT (TASK-PRICE-002) |
| #12 secrets qua Vault | - (§6/§10) | TASK-INFRA-003 |

## §4 - Kết luận

Mọi mệnh đề normative có AC + test/artefact tương ứng; không mệnh đề mồ côi. Schema khớp DATA-MODEL, ranh giới điều phối / ghi DB / lấy dữ liệu per-sàn rõ. Một khiếm khuyết typography đã sửa surgical trong lượt này. Score = 10/10. Verdict: PASS.

---

*Audit độc lập TASK-SCRAPE-001 - hết.*
