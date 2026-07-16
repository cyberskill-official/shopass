---
id: NFR-AFFIL-001
title: "Compliance affiliate - mô hình hợp lệ DUY NHẤT: user chủ động bấm -> deep link disclosure; KHÔNG cookie-stuffing kiểu Honey; KHÔNG extension scraping gắn affiliate; tuân ToS Shopee + Chrome Web Store policy"
module: AFFIL
category: compliance
priority: MUST
verification: T
phase: P2
slo: "100% affiliate link/click bắt nguồn từ hành động người dùng + có disclosure; 0 cookie-stuffing/dropping/auto-redirect; tuân ToS Shopee (cấm robots/automated query, cấm scraping; cookie chỉ set khi link hiển thị và user click voluntarily and consciously) + Chrome Web Store policy (cập nhật 3/2025, thực thi 10/06/2025)"
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_tasks: [TASK-AFFIL-001, TASK-AFFIL-002, TASK-AFFIL-003, TASK-AFFIL-004, TASK-EXT-003, TASK-TRUST-001]
source: "docs/... §4.2 (affiliate reality-check: Shopee ToS cấm robots/automated query/scraping, cookie chỉ khi link hiển thị + user click voluntarily and consciously; cấm cookie dropping/pop-under/auto-redirect/forced install; bài học PayPal Honey; Chrome Web Store policy 3/2025 thực thi 10/06/2025), §5.2 (rủi ro Chrome gỡ extension), §5.4 (minh bạch là moat niềm tin)"
---

## §1 - Statement (BCP-14 normative)

1. Mô hình affiliate hợp lệ DUY NHẤT của SănDeal **MUST** là: người dùng chủ động bấm "Mua qua SănDeal" -> hệ tạo deep link affiliate hiển thị rõ ràng (URL đích thấy được) kèm disclosure (SănDeal có thể nhận hoa hồng). Mọi affiliate link/click **MUST** bắt nguồn từ một hành động người dùng tại thời điểm đó (TASK-AFFIL-002).
2. Hệ **MUST NOT** cookie-stuffing hay cookie-dropping: KHÔNG tự set/sửa/chèn cookie affiliate trên domain sàn ở nền (kiểu PayPal Honey). Cookie last-click (nếu có) chỉ được trang sàn set khi user (đã bấm) mở deep link - không phải do SănDeal nhét.
3. Hệ **MUST NOT** dùng pop-under, auto-redirect nền, hay forced-install để gắn affiliate (theo Shopee ToS + Chrome Web Store policy 3/2025).
4. Hệ **MUST NOT** dùng extension scraping để gắn affiliate, và **MUST NOT** dùng robots/automated query/automated scraping trên sàn cho mục đích affiliate (Shopee ToS cấm "use robots or other automated query tools" và "any automated means or form of scraping, or other data extraction methods").
5. Mọi response tạo affiliate link **MUST** kèm disclosure không rỗng + URL đích hiển thị (TASK-AFFIL-002 #6/#10); che giấu đích hoặc thiếu disclosure là vi phạm.
6. Hệ **MUST** có guardrail tự động (CI gate) khẳng định không có hành vi bị cấm ở §1 #2-#4 (TASK-AFFIL-004): vi phạm làm đỏ build, không merge được.
7. Cookie affiliate **MUST** tuân điều kiện ToS Shopee: chỉ hợp lệ khi link hiển thị và user click "voluntarily and consciously" - đúng những gì TASK-AFFIL-002 thực thi (cờ user_initiated + hiển thị target_url).
8. Hệ **MUST** tuân Chrome Web Store policy (cập nhật 3/2025, thực thi 10/06/2025): cấm chèn affiliate link/code/cookie khi không mang lại lợi ích trực tiếp (giảm giá/cashback) + bắt buộc disclosure + bắt buộc hành động người dùng. Extension open-source (TASK-TRUST-001) là bằng chứng kiểm chứng được.

## §2 - Vì sao ràng buộc này

Vụ PayPal Honey là kịch bản tồn vong, không phải rủi ro lý thuyết: MegaLag phơi bày Honey thay cookie affiliate của creator ở nền (cơ chế "Selective Stand Down"), Honey mất khoảng 3 triệu trong khoảng 20 triệu user trong hai tuần, Google cập nhật chính sách Chrome Web Store tháng 3/2025 (thực thi 10/06/2025) cấm chèn affiliate khi không có lợi ích trực tiếp + bắt buộc user-action + disclosure, rồi Rakuten gỡ Honey 12/01/2026, Impact.com đình chỉ 17/01/2026, Awin ngừng thanh toán 21/01/2026. SănDeal là extension đọc cookie phiên sàn của chính người dùng - dễ bị nghi y hệt Honey. Toàn bộ định vị sản phẩm (§5.4) đặt cược vào minh bạch như một moat niềm tin: nếu SănDeal làm dù chỉ một hành vi kiểu Honey, nó vừa bị Chrome gỡ (mất kênh phân phối), vừa bị network đình chỉ (mất dòng doanh thu affiliate), vừa mất chính moat niềm tin mà nó dựa vào để khác BeeCost và phần còn lại. Đây là lý do mô hình affiliate hợp lệ phải là DUY NHẤT một đường - user chủ động bấm + disclosure - và mọi đường khác phải bị đóng ở mức kỹ thuật, không chỉ bằng lời hứa. Đồng thời ToS của chính ba sàn cấm robots/automated query/scraping cho affiliate; tuân chặt giảm cớ để sàn gửi C&D (§5.2).

## §3 - Đo lường (measurement)

- Counter `affiliate_link_created_total` vs `affiliate_link_rejected_total{reason="not_user_initiated"}` (TASK-AFFIL-002): tỷ lệ link có cờ user-initiated phải là 100% (mọi link đã tạo đều qua cờ).
- Khẳng định "mọi response link kèm disclosure": assertion `AssertSingleAffiliatePath` (TASK-AFFIL-004) + đếm response thiếu disclosure = 0.
- Guardrail CI: số kiểm tra hành vi cấm chạy + số vi phạm chặn (TASK-AFFIL-004 #12) - vi phạm bị chặn trước merge.
- Đếm `affiliate_click` không gắn user_id hợp lệ = 0 (mọi click có chủ thể user-initiated).
- Kiểm manifest extension (TASK-AFFIL-004 #4): 0 quyền `cookies` cho host sàn, 0 `webRequestBlocking` sửa redirect.
- Đối chiếu checklist Chrome Web Store policy (TASK-AFFIL-004 #8): mỗi mục policy có một test xanh.

## §4 - Verification

- Compliance gate test (T): bộ guardrail TASK-AFFIL-004 chạy như CI gate - bơm từng hành vi cấm (cookie-stuffing, pop-under, auto-redirect, link không user gesture, link thiếu disclosure, cửa affiliate thứ hai) -> mỗi cái làm đỏ build.
- User-initiated test (T): `POST /v1/affiliate/link` thiếu cờ `user_initiated` -> 400 và 0 `affiliate_click` ghi (TASK-AFFIL-002).
- Disclosure test (T): mọi response link `200` chứa disclosure không rỗng + target_url domain sàn.
- Manifest audit (T): manifest có `cookies`/`webRequestBlocking` -> guardrail đỏ; manifest tối thiểu -> xanh.
- ToS-mapping review: ánh xạ từng điều cấm của Shopee ToS (robots/automated query, scraping, cookie dropping, pop-under, auto-redirect, forced install) tới một guardrail/task thực thi.
- Open-source proof (TASK-TRUST-001): mã extension công khai + reproducible build cho phép bên thứ ba kiểm chứng không cookie-stuffing.

## §5 - Xử lý khi vi phạm

- Phát hiện một affiliate link/click không bắt nguồn từ user action -> sev-1; dừng đường tạo link, điều tra, vá ngay (đây là rủi ro tồn vong, không phải lỗi thường).
- Guardrail CI bị tắt/bỏ qua (hành vi cấm lọt merge) -> sev-1; khôi phục gate bắt buộc, audit lại các commit từ lúc tắt.
- Response link thiếu disclosure/giấu đích -> sev-2; vá để mọi response kèm disclosure + target_url; rà soát client hiển thị.
- Manifest xuất hiện quyền cookie/redirect host sàn -> sev-2; gỡ quyền, kiểm vì sao được thêm, principle of least privilege.
- Network cảnh báo/đình chỉ do nghi hành vi cấm -> sev-1; làm việc với network, dùng open-source + log postback (TASK-AFFIL-003) làm bằng chứng tuân thủ.
- Phát hiện dùng scraping/automated query gắn affiliate -> sev-1; ngừng ngay (vi phạm ToS sàn), tách bạch scraping giá (read-only, không affiliate) khỏi affiliate (user-initiated).

---

*Hết NFR-AFFIL-001.*
