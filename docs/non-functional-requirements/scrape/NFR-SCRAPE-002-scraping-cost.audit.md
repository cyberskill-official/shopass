---
nfr_id: NFR-SCRAPE-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: nfr-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tái thẩm độc lập từ file NFR hiện tại. NFR-SCRAPE-002 đặt SLO chi phí định lượng: tổng biến phí scraping (proxy + headless farm) <= ~0,1-0,2 USD/user/tháng ở cả mốc 1.000 và 100.000 users, cùng cơ chế giảm chi phí cận biên qua chia sẻ SKU. 7 mệnh đề §1 đều đo được (`proxy_cost_usd_total`, `sku_shared_ratio`, skip-rate delta-only, phân bổ theo tier proxy) và verify (cost-model/sharing/tiering/cost-guard test). Số khớp §4.1 nguồn (proxy ~$30-60 + farm ~$20-40 /1.000 users -> ~0,1-0,2 USD/user). related_tasks (TASK-SCRAPE-001/003/004/005, TASK-PRICE-002) đều resolve. Đạt 10/10.

## §2 - Findings (đã kiểm)

Kiểm frontmatter: id=NFR-SCRAPE-002 khớp tên file; category=cost, priority=MUST, verification=T, phase=P1; slo định lượng gắn cả hai mốc quy mô; source §4.1/§3.3. Đạt.

Kiểm số nguồn §4.1: line 306-310 cho proxy ~$30-60, headless farm ~$20-40 /1.000 users, tổng ~$100-200 -> ~0,1-0,2 USD/user. §1 #1 dùng đúng dải này. line 312 xác nhận cận biên giảm ở 100.000 users nhờ chia sẻ SKU - khớp §1 #2/#7. Không lệch.

Kiểm §1: 7 clause BCP-14. #2 chia sẻ SKU (quét một lần, không per-user). #3 ánh xạ 4 đòn bẩy chi phí (scan-tiering, delta-only, proxy-tiering+cost-guard, pacing). #5 cost-guard van an toàn (trần ngân sách/ngày, hạ tier khi gần trần). #6 ưu tiên đường rẻ (JSON trước farm). Đo được, không mơ hồ.

Kiểm §3 đo lường: `proxy_cost_usd_total`, `proxy_gb_used_total` per tier, chi phí farm phân bổ, `sku_shared_ratio`, snapshot_skipped vs written, phân bổ theo tier (enterprise/mid/budget). Cụ thể.

Kiểm §4 verification: cost-model test ở 1.000 và 100.000 users -> <= ~0,2 USD/user; sharing test (N user 1 SKU -> quét 1 lần); tiering test (enterprise chỉ TikTok/Lazada); cost-guard test (80% -> hạ tier, chạm trần -> dừng cold). Mỗi clause khóa có test.

Kiểm §5: chi phí vượt §4.1 sev-2, sku_shared_ratio thấp sev-3, enterprise lạm dụng sev-3, cost-guard không đặt trần sev-2, farm gọi quá thường sev-3. Hợp lý.

Kiểm phối hợp NFR-PRICE-001: delta-only skip-rate >=80% giảm cả chi phí ghi/storage - nhất quán giữa hai NFR. Đạt.

Kiểm typo: prose ASCII thuần, tiếng Việt đủ dấu, không từ cấm. Không sửa gì.

## §3 - Kết luận

SLO đo được, verify được, gắn số §4.1 và cơ chế chia sẻ SKU + 4 đòn bẩy chi phí, cost-guard là van an toàn cuối có test. Không tìm thấy defect cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập NFR-SCRAPE-002.*
