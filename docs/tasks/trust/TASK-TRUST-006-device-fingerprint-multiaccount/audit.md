---
fr_id: TASK-TRUST-006
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. TASK-TRUST-006 đặc tả device fingerprint bảo toàn riêng tư + phát hiện multi-account. Trục: hash một chiều (SHA-256 + salt server, không thuộc tính thô, CHỈ gửi hash) + chỉ ghi cạnh `device` vào đồ thị TASK-TRUST-004; multi-account lộ khi nhiều tài khoản chia sẻ thiết bị. Nguyên tắc: chỉ TÍN HIỆU, KHÔNG auto-khóa (chia sẻ thiết bị gia đình/máy net hợp lý); mục đích đơn nhất chống gian lận có cơ sở PDPL + minh bạch trong chính sách. 12 mệnh đề §1 (priority SHOULD, cấp P3). device_fingerprint khớp DATA-MODEL.md (device_hash one-way + salt server, user_id, first_seen/last_seen, PK(device_hash, user_id)). PDPL P đạt. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; priority SHOULD + phase P3 hợp lệ; key đủ; depends_on=[TASK-TRUST-004]. Pass.
- ISS-002 (contract D + privacy): device_fingerprint DDL khớp DATA-MODEL (device_hash TEXT one-way, FK user_id, PK(device_hash,user_id) idempotent); KHÔNG thuộc tính thô/PII ngoài hash + FK nội bộ (§1 #9). hash.ts chỉ gửi {device_hash}; salt server xoay được (Vault TASK-INFRA-003). Pass.
- ISS-003 (PDPL P): mục đích DUY NHẤT chống gian lận, đơn mục đích PDPL (§1 #3, DEC-TRUST-27); minh bạch trong chính sách dữ liệu TASK-TRUST-002 (§1 #8); KHÔNG dùng cho theo dõi/quảng cáo. Đúng nguyên tắc. Pass.
- ISS-004 (AC/test E,F): 12 AC; test fingerprint-no-pii.test.ts (chỉ device_hash, không UA/font/canvas; salt xoay đổi hash), TestDevice_SharedHash_WritesEdge/SoloUser_NoEdge/SharedIsSignal_NotBan (KHÔNG auto-khóa). Pass.
- ISS-005 (typography O): dấu tiếng Việt trong comment TS/SQL §3 (code block, scoped out); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 hash, không thô | #1,#2 | hash.ts + fingerprint-no-pii.test.ts |
| #2 salt server xoay được | #2 | deviceHash(serverSalt) |
| #3 mục đích đơn + PDPL | #10 | schema comment + chính sách |
| #4 cạnh device vào đồ thị | #4 | device.go AddEdge + TestDevice_SharedHash_WritesEdge |
| #5 engine tiêu thụ | #5 | graph.go nhận device |
| #6 chỉ tín hiệu | #6 | TestDevice_SharedIsSignal_NotBan |
| #7 ngưỡng cấu hình | #7,#8 | SharedAccountThreshold |
| #8 minh bạch chính sách | #10 | TASK-TRUST-002 mô tả |
| #9 không PII | #3 | device_fingerprint schema |
| #10 idempotent | #11 | PK(device_hash,user_id) + upsert |
| #11 solo không cạnh | #9 | TestDevice_SoloUser_NoEdge |
| #12 metric | #12 | device_shared_accounts_total |

## §4 - Kết luận

Hash một chiều + mục đích đơn nhất + minh bạch giữ fingerprint là công cụ phòng vệ không giám sát; chỉ tín hiệu (không auto-khóa) nhất quán nguyên tắc TRUST; schema khớp DATA-MODEL; priority SHOULD hợp lý. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-TRUST-006.*
