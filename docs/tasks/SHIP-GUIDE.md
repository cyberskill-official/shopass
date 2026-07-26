# SHIP-GUIDE.md - SănDeal (conventions build cho agent triển khai)

Hướng dẫn cho agent triển khai (implementing agent) build các Task của SănDeal. Đọc file này trước khi chạm vào bất kỳ task nào.

Quan hệ với AGENTS.md: file `AGENTS.md` ở gốc repo là giao thức memory của CyberOS (Layer-1 Memory Protocol - kích hoạt BRAIN/`.cyberos-memory/`), KHÔNG chứa conventions build. File SHIP-GUIDE.md này là conventions + bất biến khi build task. Hai file bổ trợ nhau: AGENTS.md lo memory, SHIP-GUIDE.md lo cách build. Theo §0.1 của AGENTS.md, lệnh user trong chat > AGENTS.md > mọi file hướng dẫn khác.

## SănDeal là gì

Nền tảng SaaS-tiện ích săn deal / theo dõi giá / tối ưu mua sắm đa sàn (Shopee + TikTok Shop + Lazada) cho Việt Nam và Đông Nam Á. Ba trụ cột: (A) browser extension MV3 "session piggyback" đọc giỏ hàng/voucher của chính người dùng + backend scraping giá quy mô lớn + thuật toán sale ảo / dự đoán đáy / tối ưu voucher; (B) free-tier tài trợ bằng affiliate + Premium 29k-79k VND/tháng; (C) minh bạch đạo đức hậu-Honey, tuân thủ PDPL (Luật 91/2025/QH15).

Tài liệu nền tảng đầy đủ: [`../TÀI LIỆU NỀN TẢNG SẢN PHẨM "SănDeal" ....md`](../). task nào cũng trích nguồn về tài liệu này qua `source_pages`.

## Nguồn sự thật (đọc theo thứ tự)

1. [`BACKLOG.md`](BACKLOG.md) - index toàn bộ 90 task theo phase -> module -> slice, với priority/status/depends_on/effort.
2. [`IMPLEMENTATION-ORDER.md`](IMPLEMENTATION-ORDER.md) - thứ tự build theo 8 layer topo (DAG đã verify acyclic + reciprocal).
3. [`DATA-MODEL.md`](DATA-MODEL.md) - schema DB hợp nhất: một table một owner task.
4. [`STATUS-REFERENCE.md`](STATUS-REFERENCE.md) - enum 10 trạng thái + vòng đời.
5. File task riêng: `<module>/FR-<MODULE>-NNN-<slug>.md` + companion `.audit.md`.
6. NFR: `../non-functional-requirements/<module>/NFR-<MODULE>-NNN-<slug>.md`.

## Chọn task tiếp theo + cách build

1. Mở IMPLEMENTATION-ORDER.md, lấy task ở layer thấp nhất mà mọi `depends_on` đã `done`. Trong cùng layer: `MUST` trước `SHOULD` trước `COULD`.
2. Lật status task `ready_to_implement` -> `implementing` (xem STATUS-REFERENCE.md).
3. Build theo frontmatter: `new_files` (file cần tạo), `sub_tasks` (chia nhỏ + ước lượng giờ), `service`/`language` (vị trí + ngôn ngữ). Tôn trọng `allowed_tools` / `disallowed_tools`.
4. Mỗi mệnh đề MUST/SHOULD ở §1 phải có test ở §5 pass và đối chiếu acceptance criteria §4. §3 cho hợp đồng API/DDL, §8 payload ví dụ, §10 failure modes.
5. Đi theo vòng đời tới `done`; cập nhật ô status trên BACKLOG.

Mỗi task đã qua audit độc lập 10/10 (file `.audit.md`, `auditor: independent`) nên đặc tả là tự chứa - không cần hỏi lại tác giả để build.

## Cấu trúc một task (engineering-spec@1)

Frontmatter YAML (id, title, module, priority, status, phase, slice, depends_on, blocks, source_pages, source_decisions, language, service, new_files, modified_files, allowed_tools, disallowed_tools, effort_hours, sub_tasks, risk_if_skipped) + thân §1 Mô tả (BCP-14 normative) / §2 Vì sao / §3 Hợp đồng API / DDL / §4 Acceptance criteria / §5 Kiểm thử / §6 Khung triển khai / §7 Phụ thuộc / §8 Payload ví dụ / §9 Câu hỏi mở / §10 Failure modes / §11 Ghi chú.

## Bất biến KHÔNG thương lượng (mọi task phải giữ)

Đây là các ràng buộc xuyên suốt. Vi phạm là lỗi cứng dù task không nhắc lại.

1. Bảo mật và PDPL. KHÔNG lưu cleartext credential. Token phiên sàn (Shopee/TikTok/Lazada) KHÔNG rời client và KHÔNG bao giờ lưu trên server - `platform_account.ext_user_ref` là định danh ẩn danh, không phải token. Mật khẩu băm argon2id định dạng PHC (lưu tham số kèm hash). Secrets trong Vault / AWS Secrets Manager. Xử lý dữ liệu cá nhân phải có consent PDPL (tự nguyện, cụ thể, đơn mục đích; im lặng khác đồng thuận); DPIA nộp trong 60 ngày; thông báo vi phạm trong 72 giờ. Chế tài PDPL tới 5% doanh thu / 10 lần lợi nhuận bất chính / 3 tỷ VND.
2. Affiliate hậu-Honey. CHỈ tạo affiliate deep link khi user CHỦ ĐỘNG bấm, hiển thị disclosure rõ. KHÔNG cookie-stuffing/dropping/auto-redirect/pop-under (Honey-style). KHÔNG dùng extension scraping để gắn affiliate. Tuân Chrome Web Store policy (cập nhật 3/2025, thực thi 10/06/2025). Cashback chỉ tính trên conversion đã confirmed + hold-then-release + delay payout.
3. Tiền tệ. Mọi cột tiền là `BIGINT` đơn vị VND, không float/numeric - tránh sai số trên phép so phần trăm (sale ảo, optimizer).
4. Giá. Ghi `price_snapshot` theo delta-only (chỉ ghi khi giá đổi). `price_snapshot` là TimescaleDB hypertable + nén + continuous aggregate `price_daily`. Đọc biểu đồ/lịch sử từ `price_daily`, không quét raw.
5. Extension MV3. Service worker ephemeral (không giữ state trong biến global - dùng `chrome.storage`); `chrome.alarms` tối thiểu 30s, không `setInterval`; tác vụ nặng đẩy backend; Offscreen API cho DOM/clipboard; `declarativeNetRequest` thay webRequest blocking. Content script CHỈ gửi productId/price/qty về backend, KHÔNG cookie/token. Ưu tiên đọc DOM render với TikTok Shop (né msToken/_signature/X-Bogus) và Lazada (Akamai).
6. Notification. Cost model ưu tiên push > email > sms. FCM HTTP v1, quota 600.000 msg/phút/project, xử lý 429 RESOURCE_EXHAUSTED bằng backoff. Đỉnh 00:00 phải flatten-the-curve (jitter + bucket theo phút). Kênh push tách theo `user_channel_token.platform`: FCM nhặt `platform IN ('android','web')`, APNs nhặt `platform='ios'`.
7. Data model. Một table một owner task (xem DATA-MODEL.md). Module khác cần một table thì `depends_on` owner và tham chiếu, KHÔNG re-create. Mở rộng bằng `ALTER TABLE` (ví dụ AUTH-001 thêm `pwd_hash` vào `app_user` lõi của INFRA-002), không redefine cột lõi.
8. Per-country gating. VN trước; MY/PH bỏ stacking voucher 2025 (freeship gộp nhóm platform); luật bảo vệ dữ liệu khác nhau (PDPL VN, PDP Indonesia, PDPA Thái Lan). Đọc CountryPolicy, mặc định restrictive (no-stack) cho nước chưa cấu hình.
9. Scraping. Bắt buộc residential proxy (datacenter vô dụng với Cloudflare/Akamai); pacing ngẫu nhiên + jitter; giám sát DOM drift. Tự động hóa xu/voucher rủi ro ban cao -> chỉ checklist nhắc nhở + auto-test mã do user khởi tạo (sleep 2.5-5s, revert, không tự chốt đơn).

## Stack theo module

- Backend service (auth, scrape, price, track, notif, cart, affil, bill, b2b, comply, trust): Go 1.25.12 + PostgreSQL 16; `price` thêm TimescaleDB; queue Redis / Kafka-Redis Streams.
  Rationale (R10, 2026-07-26): keep the 1.25 line (code/CI/Dockerfile already there; do not pin back to 1.22). Pin patch `1.25.12` via `toolchain go1.25.12` in every `go.mod`, CI `setup-go`, and `deploy/Dockerfile.go` so govulncheck stays green on stdlib CVEs.
- ML (deal: dự đoán đáy giá): Python (Prophet baseline -> LightGBM) + feature store.
- Extension: TypeScript + Chrome Manifest V3.
- Web: Next.js + TypeScript (App Router) + GraphQL BFF.
- Mobile (giai đoạn P3): React Native hoặc Flutter.
- Hạ tầng: API Gateway/BFF, Vault, OTel + Prometheus + Grafana, residential proxy + Playwright headless farm.

## Quy ước viết

Văn xuôi trong task/spec viết tiếng Việt; ID, code, SQL, tên API, tên cột, từ khóa kỹ thuật giữ tiếng Anh. Chỉ ký tự bàn phím chuẩn: dấu gạch nối thường (không em-dash/en-dash), dấu nháy thẳng, ba dấu chấm cho dấu lược, "->" thay mũi tên, ">=" / "<=" thay ký hiệu Unicode. Dấu `§` được giữ làm quy ước đánh mục. Không emoji trong văn xuôi. Code block không bị các quy tắc này ràng buộc.

## Khi sửa một task

Nếu phải đổi `depends_on` của một task, đồng bộ lại: cập nhật cột Depends on trong BACKLOG.md, tính lại `blocks` (nghịch đảo của depends_on) cho mọi task, và regenerate IMPLEMENTATION-ORDER.md. Nếu đổi acceptance criteria đủ nhiều để dịch số AC, cập nhật bảng traceability trong file `.audit.md` tương ứng.
