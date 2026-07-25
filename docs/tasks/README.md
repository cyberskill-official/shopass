# Shopass - Tasks (đọc đầu tiên)

Thư mục này là backlog Task (task) của Shopass: 90 task + 10 NFR đặc tả toàn bộ sản phẩm, viết theo workflow `task` của CyberOS (engineering-spec@1). Trạng thái: v0.2.0 SHIP-READY - mọi task đã qua audit độc lập 10/10, DAG phụ thuộc sạch, data model nhất quán. Đây là spec để xây, chưa có code; mục tiêu là một người hoặc một agent đọc xong là build được.

README này dành cho cả người mới và agent triển khai. Đọc hết phần "Bắt đầu từ đâu" rồi nhảy tới phần phù hợp.

## Bắt đầu từ đâu

Nếu bạn là người mới (muốn hiểu sản phẩm và backlog):

1. Đọc phần "Shopass là gì" ngay dưới.
2. Đọc tài liệu nền tảng đầy đủ: [`../TÀI LIỆU NỀN TẢNG SẢN PHẨM "SănDeal" — PRD + SRS + CHIẾN LƯỢC KỸ THUẬT : KINH DOANH : RỦI RO.md`](../TÀI LIỆU NỀN TẢNG SẢN PHẨM "SănDeal" — PRD + SRS + CHIẾN LƯỢC KỸ THUẬT : KINH DOANH : RỦI RO.md) (PRD + SRS + chiến lược kỹ thuật/kinh doanh/rủi ro).
3. Đọc [`BACKLOG.md`](BACKLOG.md) để thấy bức tranh tổng: 4 phase, 16 module, 90 task.
4. Mở một task mẫu để xem một spec trông thế nào: [`price/TASK-PRICE-002-price-snapshot-hypertable/spec.md`](price/TASK-PRICE-002-price-snapshot-hypertable/spec.md).

Nếu bạn là agent triển khai (sắp build code):

1. Đọc `AGENTS.md` ở gốc repo (giao thức memory CyberOS - kích hoạt BRAIN) nếu môi trường dùng memory.
2. Đọc [`SHIP-GUIDE.md`](SHIP-GUIDE.md): conventions + 9 bất biến KHÔNG thương lượng. Bắt buộc.
3. Mở [`IMPLEMENTATION-ORDER.md`](IMPLEMENTATION-ORDER.md), lấy task ở layer thấp nhất mà mọi `depends_on` đã `done`.
4. Đọc file task đó, build theo `new_files` + `sub_tasks`, chạy test §5, đối chiếu acceptance criteria §4, lật status tới `done`.

## Shopass là gì

Nền tảng SaaS-tiện ích săn deal / theo dõi giá / tối ưu mua sắm đa sàn (Shopee + TikTok Shop + Lazada) cho Việt Nam và Đông Nam Á. Ba trụ cột: (A) browser extension Manifest V3 "session piggyback" đọc giỏ hàng/voucher của chính người dùng + backend scraping giá quy mô lớn + thuật toán sale ảo / dự đoán đáy giá / tối ưu voucher; (B) free-tier tài trợ bằng affiliate + Premium 29k-79k VND/tháng; (C) minh bạch đạo đức hậu-Honey, tuân thủ PDPL (Luật 91/2025/QH15). Moat: dữ liệu lịch sử giá tích lũy + so sánh chéo 3 sàn + niềm tin.

## Bản đồ tài liệu (đọc cái nào khi nào)

| File | Là gì | Đọc khi |
|---|---|---|
| [`BACKLOG.md`](BACKLOG.md) | Index nguồn-sự-thật: 4 phase -> 16 module -> slice -> 90 task, kèm priority/status/depends_on/effort + tổng quan + cổng thoát phase + risk register | Muốn bức tranh tổng, hoặc tra một task thuộc đâu |
| [`SHIP-GUIDE.md`](SHIP-GUIDE.md) | Conventions build + 9 bất biến không thương lượng + stack theo module + quy ước viết | Trước khi build bất kỳ task nào |
| [`IMPLEMENTATION-ORDER.md`](IMPLEMENTATION-ORDER.md) | Thứ tự build theo 8 layer topo (DAG acyclic + reciprocal) + critical path MVP | Chọn task tiếp theo để build |
| [`DATA-MODEL.md`](DATA-MODEL.md) | Schema DB hợp nhất: ~50 bảng, mỗi bảng một owner task, cột khóa + FK | Viết code đụng DB, kiểm tham chiếu bảng/cột |
| [`STATUS-REFERENCE.md`](STATUS-REFERENCE.md) | Enum 10 trạng thái + vòng đời task + cách lật status | Khi đổi status một task |
| `../../AGENTS.md` (gốc repo) | Giao thức memory CyberOS (BRAIN). KHÔNG phải conventions build | Môi trường có dùng memory CyberOS |
| File task/NFR riêng | Đặc tả chi tiết từng yêu cầu | Khi build chính task đó |

## Bố cục thư mục

```
shopass/
  AGENTS.md                          # (gốc) giao thức memory CyberOS - KHÔNG phải conventions build
  docs/
    TÀI LIỆU NỀN TẢNG ... SănDeal.md  # tài liệu nguồn (PRD + SRS + chiến lược)
    tasks/
      README.md                      # file này
      BACKLOG.md                     # index 90 task
      SHIP-GUIDE.md                  # conventions + bất biến
      IMPLEMENTATION-ORDER.md        # thứ tự build 8 layer
      DATA-MODEL.md                  # schema DB hợp nhất
      STATUS-REFERENCE.md            # enum status + vòng đời
      <module>/                      # 16 module (xem bảng dưới)
        FR-<MODULE>-NNN-<slug>.md     # đặc tả task (engineering-spec@1)
        FR-<MODULE>-NNN-<slug>.audit.md  # audit độc lập đi kèm (10/10)
    non-functional-requirements/
      <module>/
        NFR-<MODULE>-NNN-<slug>.md    # ràng buộc phi chức năng + SLO
        NFR-<MODULE>-NNN-<slug>.audit.md
```

## 16 module (90 task qua 4 phase)

| Module | task | Phase | Ngôn ngữ/stack | Lo việc gì |
|---|---:|---|---|---|
| infra | 5 | P0 | Go + hạ tầng | API Gateway/BFF, data-model foundation, secrets/Vault, observability, per-country config |
| auth | 5 | P1 | Go + Postgres | app_user, JWT, liên kết tài khoản sàn (no token), social login, vòng đời tài khoản |
| scrape | 8 | P1-P2 | Go + Playwright | orchestrator, adapter 3 sàn, residential proxy, anti-fingerprint, delta-only, DOM-monitoring |
| price | 5 | P1 | Go + TimescaleDB | tracked_product, price_snapshot hypertable, lịch sử giá, so sánh chéo sàn, canonical_key |
| ext | 8 | P1-P2 | TypeScript + MV3 | scaffold MV3, content script 3 sàn, tối thiểu hóa dữ liệu, offscreen, sync, consent UI |
| track | 4 | P1 | Go + Postgres | track product, wishlist, alert_rule, engine kích hoạt alert |
| deal | 6 | P1-P2 | Python + Go | sale ảo, cold-start, chart data, dự đoán đáy (Prophet -> LightGBM), nightly scoring |
| notif | 7 | P1-P2 | Go + Kafka/Redis | schema + routing, FCM, fan-out + DLQ, midnight-spike, APNs, email, SMS VN |
| web | 5 | P1 | Next.js + TS | scaffold, SEO landing, biểu đồ giá, UI wishlist/alert, GraphQL BFF |
| comply | 8 | P1-P3 | Go + Postgres | PDPL consent, DPIA/TIA, DSAR, breach 72h, no-cleartext, per-country gating, SEA, e-commerce law |
| trust | 6 | P1-P3 | Go + extension | open-source, data-minimization, security audit, anti-fraud, attribution-gaming, device fingerprint |
| cart | 6 | P2 | Go + TS | voucher catalog, cart snapshot, optimizer, per-country stacking, auto-test mã, checklist xu |
| affil | 5 | P2-P3 | Go + Postgres | affiliate tracking, user-initiated deeplink, network integration, Honey-avoidance, cashback |
| bill | 5 | P2 | Go + Postgres | subscription, payment gateway (MoMo/ZaloPay/VNPay/VietQR), reconciliation, referral, upgrade gating |
| b2b | 4 | P3 | Go + TimescaleDB | dữ liệu xu hướng ẩn danh (k-anonymity), báo cáo, seller analytics, premium API |
| mobile | 3 | P3 | React Native/Flutter | scaffold + push, tracking + checkout assistant, deep-link virality |

Phân bố phase: P0 = 5 task (~33h), P1 = 46 task (~303h), P2 = 25 task (~177h), P3 = 14 task (~114h). Tổng ~627h (~16 person-week thuần code).

## Đọc một task thế nào

Mỗi task là một file engineering-spec@1: frontmatter YAML + 11 mục thân. File mẫu chuẩn (gold standard): [`price/TASK-PRICE-002-price-snapshot-hypertable/spec.md`](price/TASK-PRICE-002-price-snapshot-hypertable/spec.md).

Frontmatter chính: `id`, `title`, `module`, `priority` (MUST/SHOULD/COULD/MAY), `status`, `phase`, `slice`, `depends_on` (task phải xong trước), `blocks` (task phụ thuộc vào cái này), `source_pages` (trỏ về tài liệu nguồn), `source_decisions` (DEC-<MODULE>-NN: quyết định thiết kế đã chốt), `language`, `service` (vị trí code), `new_files` (file cần tạo), `modified_files`, `allowed_tools`/`disallowed_tools`, `effort_hours`, `sub_tasks` (chia nhỏ + giờ), `risk_if_skipped`.

Thân 11 mục:

- §1 Mô tả (BCP-14 normative) - mệnh đề MUST/SHOULD đánh số, mỗi mệnh đề kiểm thử được. Đây là hợp đồng.
- §2 Vì sao thiết kế này - lý do cho người đọc.
- §3 Hợp đồng API / DDL - schema SQL, types, handler thật.
- §4 Acceptance criteria - map 1:1 với mệnh đề §1.
- §5 Kiểm thử - test bodies thật, mỗi mệnh đề kiểm thử được có test.
- §6 Khung triển khai - thứ tự dựng.
- §7 Phụ thuộc - task + thư viện.
- §8 Payload ví dụ.
- §9 Câu hỏi mở.
- §10 Failure modes inventory - bảng lỗi + phát hiện + khắc phục.
- §11 Ghi chú.

Mỗi task có file `.audit.md` đi kèm (`auditor: independent`, score 10/10) ghi lại verdict + traceability §1 -> AC -> artefact.

## Quy trình build cho agent (step-by-step)

1. Đọc SHIP-GUIDE.md (conventions + 9 bất biến). Đây là điều kiện tiên quyết.
2. Mở IMPLEMENTATION-ORDER.md. Lấy task khả thi tiếp theo = task ở layer thấp nhất chưa `done` mà mọi `depends_on` đã `done`. Trong cùng layer: `MUST` trước `SHOULD` trước `COULD`. Layer 0 là 4 root: TASK-EXT-001, TASK-INFRA-001, TASK-INFRA-002, TASK-INFRA-003.
3. Lật status task từ `ready_to_implement` sang `implementing` (xem STATUS-REFERENCE.md), cập nhật ô status trên BACKLOG.
4. Build theo frontmatter: tạo đúng `new_files`, làm theo `sub_tasks`, đặt code ở `service`, tôn trọng `allowed_tools`/`disallowed_tools`.
5. Viết test §5; mỗi mệnh đề MUST/SHOULD ở §1 phải có test pass; đối chiếu acceptance criteria §4.
6. Đi tiếp vòng đời: `implementing` -> `ready_to_review` -> `reviewing` -> `ready_to_test` -> `testing` -> `done`. Cập nhật ô status trên BACKLOG mỗi bước.
7. Nếu một bước fail hoặc gặp blocker, status rớt về `ready_to_implement` (xem STATUS-REFERENCE §1.3).

Critical path MVP (xây trước để chạy được vòng giá trị lõi): TASK-INFRA-002 -> TASK-PRICE-001 -> TASK-PRICE-002 -> TASK-DEAL-001/003 -> TASK-WEB-003. Song song: TASK-SCRAPE-001/002/005 (bắt đầu sớm để tích lũy 90 ngày dữ liệu cho sale ảo); TASK-EXT-001/002/003.

## 9 bất biến KHÔNG thương lượng (tóm tắt - chi tiết ở SHIP-GUIDE.md)

Mọi task phải giữ, dù task không nhắc lại: (1) no-cleartext + token phiên sàn không rời client/không lưu server + argon2id + PDPL consent/DPIA/breach 72h; (2) affiliate chỉ user-initiated + disclosure, né Honey/cookie-stuffing, tuân Chrome policy 10/06/2025; (3) tiền BIGINT VND, không float; (4) giá ghi delta-only, price_snapshot là TimescaleDB hypertable, đọc biểu đồ từ price_daily; (5) extension MV3 (SW ephemeral, alarms >=30s, no global state, chỉ gửi productId/price/qty); (6) notification push>email>sms, FCM 600k/phút, flatten-the-curve, tách platform; (7) một table một owner task, mở rộng bằng ALTER; (8) per-country gating (VN trước, MY/PH no-stack 2025); (9) scraping residential proxy + pacing, xu/voucher chỉ checklist + auto-test mã user-initiated.

## Đọc BACKLOG thế nào

BACKLOG tổ chức theo phase (P0 Nền tảng -> P1 MVP -> P2 Mở rộng -> P3 Tăng trưởng), trong mỗi phase theo module, trong mỗi module theo slice (đơn vị ship gọn; slice 1 luôn là bề mặt tối thiểu). Mỗi dòng là một task với priority (BCP-14: MUST/SHOULD/COULD/MAY), status, depends_on, effort. §1 cho tổng quan, §7 cho thứ tự build, §8 cho tham chiếu chéo risk register.

## Quy ước

- Ngôn ngữ: văn xuôi tiếng Việt; ID, code, SQL, tên API, tên cột, từ khóa kỹ thuật giữ tiếng Anh.
- Typography: chỉ ký tự bàn phím chuẩn - dấu gạch nối thường (không em-dash), dấu nháy thẳng, ba dấu chấm cho dấu lược, "->" thay mũi tên, ">=" / "<=" thay ký hiệu Unicode. Dấu `§` được giữ làm quy ước đánh mục. Không emoji trong văn xuôi. Code block không bị ràng buộc các quy tắc này.
- Status enum (10): `draft | ready_to_implement | implementing | ready_to_review | reviewing | ready_to_test | testing | done | on_hold | closed`. Mọi task khởi tạo ở `ready_to_implement`.

## Sửa một task an toàn

Nếu đổi `depends_on` của một task: cập nhật cột Depends on trong BACKLOG.md, tính lại `blocks` (nghịch đảo của depends_on) cho mọi task, regenerate IMPLEMENTATION-ORDER.md. Nếu đổi acceptance criteria đủ để dịch số AC: cập nhật bảng traceability trong file `.audit.md` tương ứng. Một table chỉ một owner task - module khác cần bảng thì `depends_on` owner và tham chiếu, không re-create.

## Thuật ngữ

- task (Task): một yêu cầu nguyên tử, kiểm thử được; một file engineering-spec@1.
- NFR: ràng buộc phi chức năng (hiệu năng, bảo mật, chi phí, độ tin cậy) kèm SLO định lượng.
- audit: file `.audit.md` chấm điểm task/NFR theo rubric; `auditor: independent` nghĩa là đã qua gate audit riêng (không tự-chấm).
- DEC: quyết định thiết kế đã chốt (DEC-<MODULE>-NN), trích trong `source_decisions`.
- module: một service hoặc bề mặt (16 cái); slice: đơn vị ship trong module; layer: bậc topo trong DAG phụ thuộc (8 bậc).
- BRAIN / memory: hệ bộ nhớ CyberOS, kích hoạt qua AGENTS.md gốc + store `.cyberos-memory/`.

## Chất lượng và nguồn gốc

Backlog dẫn xuất từ tài liệu nền tảng SănDeal v1.0 (16/06/2026), áp dụng workflow task của CyberOS. Mỗi task đã qua: tác giả -> audit độc lập (gate riêng, không tự-chấm) -> review chéo module -> kiểm tra DAG/data-model/typography bằng script. 90 task + 10 NFR đều PASS 10/10. DAG acyclic, 0 lỗi reciprocity, 0 đụng độ schema. `source_pages` trong mỗi task trỏ về mục tương ứng của tài liệu nguồn.

## Câu hỏi thường gặp (người mới)

- Code ở đâu? Phần lõi đã có trong `services/`, `extension/`, `web/`. `mobile/` chưa có trong repo (P3; xem `docs/TASK-COVERAGE.md`). Spec nằm ở thư mục này; agent build theo `new_files` của từng task.
- Build cái gì trước? Layer 0 trong IMPLEMENTATION-ORDER: TASK-EXT-001 + TASK-INFRA-001/002/003.
- Một task có "đủ" để build không? Có. Mỗi task tự chứa (frontmatter + §1-§11) và đã qua audit độc lập 10/10, không cần hỏi lại tác giả.
- AGENTS.md và SHIP-GUIDE.md khác gì? AGENTS.md (gốc) là giao thức memory CyberOS; SHIP-GUIDE.md là conventions build Shopass. Hai thứ tách biệt, bổ trợ nhau.
