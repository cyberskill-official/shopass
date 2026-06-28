---
nfr_id: NFR-SCRAPE-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: nfr-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tái thẩm độc lập từ file NFR hiện tại. NFR-SCRAPE-001 gắn resilience anti-bot trực tiếp vào bảng xếp hạng rủi ro ban §3.9: scraping read-only Shopee Medium-High / TikTok High / Lazada Medium-High; đọc giỏ qua extension Low; tự động xu/voucher High. 7 mệnh đề §1 đều có cơ chế đo (`scrape_ban_rate`, `scrape_success_rate`, `adapter_health_state`, MTTR adapter) và verify (resilience/tiering/DOM-drift/risk-boundary/multi-platform test). Bảng §1 #1 khớp đúng §3.9 nguồn (line 295-297). related_frs (FR-SCRAPE-001..008) đều resolve. Đạt 10/10.

## §2 - Findings (đã kiểm)

Kiểm frontmatter: id=NFR-SCRAPE-001 khớp tên file; category=reliability, priority=MUST, verification=T, phase=P1; source §3.9/§3.2/§5.2. Đạt.

Kiểm bảng rủi ro §1 #1 vs nguồn §3.9: Shopee Medium-High, TikTok High, Lazada Medium-High cho scraping read-only; cả 3 Low cho đọc giỏ extension; cả 3 High cho xu/voucher. Khớp từng ô. Không lệch.

Kiểm §1 mệnh đề khác: #2 ánh xạ mức rủi ro -> lớp chống ban cụ thể (TikTok/Lazada bắt buộc enterprise proxy). #4 tách rõ đường extension Low-risk first-party khỏi hạ tầng scraping. #5 MUST NOT tự động xu/voucher (ranh giới tự đặt). #6 phục hồi DOM-drift trong MTTR giới hạn. Đo được.

Kiểm §3 đo lường: ban-rate per platform, success-rate, `proxy_ip_banned_total`, `adapter_health_state` (healthy/degraded/broken), MTTR adapter; phân biệt `parse_fail` (DOM đổi) với `challenge` (anti-bot). Cụ thể và chẩn đoán đúng nguyên nhân.

Kiểm §4 verification: resilience test đo ban-rate dưới ngưỡng, tiering test xác nhận SelectTier đúng (enterprise cho TikTok/Lazada), DOM-drift recovery test đo MTTR, risk-boundary test (extension không qua scraping; không code tự-click xu), multi-platform degradation test. Mỗi clause khóa có test.

Kiểm §5: ban-rate vượt sev-2, broken quá MTTR sev-2, đốt IP bất thường sev-3, code tự-click xu -> chặn merge (ranh giới compliance cứng), C&D toàn diện sev-1 chiến lược. Hợp lý.

Kiểm typo: prose ASCII thuần ("->", ">="), tiếng Việt đủ dấu, không từ cấm. Bảng markdown hợp lệ. Không sửa gì.

## §3 - Kết luận

SLO đo được (ban-rate, success-rate, health-state, MTTR), verify được, gắn cơ chế FR-SCRAPE-003/004/005/006 và bảng rủi ro §3.9. Ranh giới Low/High (extension vs scraping vs xu-voucher) rõ và có test. Không tìm thấy defect cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập NFR-SCRAPE-001.*
