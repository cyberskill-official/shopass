---
nfr_id: NFR-AFFIL-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: nfr-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tái thẩm độc lập từ file NFR hiện tại. NFR-AFFIL-001 đặt ràng buộc compliance định lượng và kiểm được: mô hình affiliate hợp lệ DUY NHẤT là user chủ động bấm -> deep link disclosure (URL đích hiển thị), và mọi hành vi kiểu Honey (cookie-stuffing/dropping, pop-under, auto-redirect, forced-install, extension-scraping-affiliate) bị cấm tuyệt đối qua guardrail CI. 8 mệnh đề §1 đều có cơ chế đo (tỷ lệ link user-initiated 100%, disclosure khác rỗng, guardrail vi phạm bị chặn, manifest audit) và verify (compliance gate test, user-initiated test, ToS-mapping). Số/sự kiện khớp §4.2 nguồn (Shopee ToS robots/scraping; cookie chỉ khi link hiển thị + click voluntarily and consciously; Chrome policy 3/2025 thực thi 10/06/2025; Honey ~3 triệu trong ~20 triệu user 2 tuần). related_tasks (TASK-AFFIL-001/002/003/004, TASK-EXT-003, TASK-TRUST-001) đều resolve. Đạt 10/10.

## §2 - Findings (đã kiểm)

Kiểm frontmatter: id=NFR-AFFIL-001 khớp tên file; category=compliance, priority=MUST, verification=T, phase=P2; slo định lượng (100% link có disclosure + user-action; 0 cookie-stuffing); source §4.2/§5.2/§5.4. Đạt.

Kiểm số/dẫn chiếu nguồn §4.2: line 326 Shopee ToS cấm "use robots or other automated query tools", "any automated means or form of scraping"; cookie chỉ khi link hiển thị + click "voluntarily and consciously"; cấm cookie dropping/pop-under/auto-redirect/forced install. line 330 Chrome Web Store policy 3/2025 thực thi 10/06/2025; Honey mất ~3 triệu trong ~20 triệu user 2 tuần; chuỗi gỡ Rakuten 12/01/2026, Impact.com 17/01/2026, Awin 21/01/2026. §1 #2/#3/#4/#7/#8 và §2 dùng đúng các dẫn chiếu này. Không lệch.

Kiểm §1: 8 clause BCP-14. #1 mô hình hợp lệ DUY NHẤT (user-initiated + disclosure). #6 guardrail CI gate đỏ build khi vi phạm (cưỡng chế kỹ thuật, không chỉ lời hứa). #7 ánh xạ trực tiếp điều kiện cookie ToS Shopee sang cờ user_initiated + hiển thị target_url. Đo được.

Kiểm §3 đo lường: counter link_created vs rejected{not_user_initiated} (tỷ lệ user-initiated 100%), AssertSingleAffiliatePath + đếm response thiếu disclosure = 0, guardrail CI số vi phạm bị chặn, manifest audit (0 quyền cookies host sàn, 0 webRequestBlocking sửa redirect). Cụ thể.

Kiểm §4 verification: compliance gate test (bơm từng hành vi cấm -> đỏ build), user-initiated test (thiếu cờ -> 400, 0 click ghi), disclosure test (mọi 200 có disclosure + target_url domain sàn), manifest audit, ToS-mapping review, open-source proof (reproducible build). Mỗi clause khóa có test.

Kiểm §5: link không user-action sev-1, guardrail bị tắt sev-1, thiếu disclosure sev-2, quyền cookie host sàn sev-2, network đình chỉ sev-1, scraping gắn affiliate sev-1. Đặt đúng sev-1 cho rủi ro tồn vong.

Kiểm typo: prose ASCII thuần, tiếng Việt đủ dấu, không từ cấm. Cụm tiếng Anh trích ToS ("voluntarily and consciously") để nguyên là trích nguồn - hợp lệ. Không sửa gì.

## §3 - Kết luận

Ràng buộc compliance đo được (user-initiated, disclosure, guardrail, manifest), verify được (gate test + ToS-mapping + open-source proof), gắn đúng ToS Shopee + Chrome policy + bài học Honey, cưỡng chế ở mức kỹ thuật qua TASK-AFFIL-002/004. Xử lý vi phạm đặt đúng sev-1 cho rủi ro tồn vong. Không tìm thấy defect cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập NFR-AFFIL-001.*
