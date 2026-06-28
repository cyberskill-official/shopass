---
nfr_id: NFR-PRICE-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: nfr-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tái thẩm độc lập từ file NFR hiện tại. NFR-PRICE-001 đặt SLO định lượng cho lớp time-series: query 90 ngày 1 SKU p95 < 500ms ở quy mô >=1 tỷ dòng (qua continuous aggregate `price_daily`), raw 7 ngày p95 < 300ms, storage biến phí <= ~0,1-0,2 USD/user/tháng, throughput ghi >=5.000 INSERT-quyết-định/s, nén >=8x. 6 mệnh đề §1 đều đo được và có verify (load/compression/throughput test). Số khớp §3.8 (mở rộng tỷ-dòng) + §4.1 (unit economics 0,1-0,2 USD). related_frs (FR-PRICE-002/003, FR-DEAL-001, FR-SCRAPE-005) đều resolve. Đạt 10/10 sau một sửa surgical.

## §2 - Findings (đã kiểm)

Kiểm frontmatter: id=NFR-PRICE-001 khớp tên file; category=scalability, priority=MUST, verification=T, phase=P1; source trỏ §3.8/§3.4/§4.1. Đạt.

ISS-FM-01 (đã sửa) - mâu thuẫn nội bộ con số storage. Frontmatter `slo` ghi `storage <= 0,1 USD/user/tháng` trong khi §1 #3 và nguồn §4.1 dùng dải `~0,1-0,2 USD`. Sửa surgical: đổi frontmatter thành `<= ~0,1-0,2 USD/user/tháng` để khớp statement và nguồn. Đây là sửa duy nhất.

Kiểm §1: 6 clause BCP-14, mỗi clause có ngưỡng (500ms/300ms/0,1-0,2 USD/8x/5.000 INSERT/18 tháng retention). #1 gắn quy mô >=1 tỷ dòng tường minh. #3 gắn cơ chế đạt (delta-only #4 + nén columnar #5). Không mơ hồ.

Kiểm §3 đo lường: histogram `price_query_duration_ms`, gauge rows/bytes nén, counter written vs delta_skipped, panel chi phí storage/user. Cụ thể.

Kiểm §4 verification: seed >=1 tỷ dòng (1 triệu SKU x 1.000 snapshot) đo p95 90 ngày, compression test >=8x, throughput test 5.000 InsertSnapshot/s với 90% no-change, reconciliation skip-rate >=80%. Mỗi clause khóa có test.

Kiểm §5: query vượt sev-3, nén thấp sev-3, storage vượt §4.1 sev-2, throughput không kịp flash sale sev-2. Hợp lý.

Kiểm khớp DATA-MODEL: `price_snapshot` hypertable chunk 7 ngày, nén sau 30 ngày, retention raw 18 tháng, segmentby product_id - khớp §1 #2/#4/#6 và frontmatter. Đạt.

Kiểm typo: prose ASCII thuần, tiếng Việt đủ dấu, không từ cấm. Sạch ngoài sửa ISS-FM-01.

## §3 - Kết luận

SLO đo được, verify được, gắn cơ chế FR-PRICE-002 (hypertable + delta-only + nén) và số nguồn §4.1. Một mâu thuẫn nội bộ frontmatter đã sửa surgical. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập NFR-PRICE-001.*
