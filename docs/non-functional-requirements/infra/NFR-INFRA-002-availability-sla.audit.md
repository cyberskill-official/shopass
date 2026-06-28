---
nfr_id: NFR-INFRA-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: nfr-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Tái thẩm độc lập từ file NFR hiện tại. NFR-INFRA-002 đặt SLO khả dụng 99,5%/tháng cho bề mặt lõi, gắn ngân sách lỗi định lượng ~3h39m và SLI success-rate (không phải ping). 6 mệnh đề §1 phủ uptime, cách đo, graceful degradation, no-SPOF, bảo trì không downtime, và chính sách error-budget. Mỗi mệnh đề khóa có verify (đo SLI production, chaos test, deploy test). Con số 99,5% khớp §3.8 nguồn; ngân sách lỗi ~3h39m đúng (0,5% của tháng trung bình ~730h). related_frs đều resolve. Đạt 10/10.

## §2 - Findings (đã kiểm)

Kiểm frontmatter: id=NFR-INFRA-002 khớp tên file; title quoted; category=availability, priority=MUST, verification=T, phase=P0; slo định lượng (99,5% + ~3h39m error budget); source §3.8. Đạt.

Kiểm §1: 6 clause BCP-14. #2 ép SLI success-rate thay vì ping - chặn đúng bẫy "cổng ping được nhưng trả 5xx". #3 graceful degradation (cagg trễ -> dữ liệu cũ + dấu hiệu). #6 error-budget policy (freeze tính năng khi cạn). Ngưỡng đo được, không mơ hồ.

Kiểm số ngân sách lỗi: 99,5% -> 0,5% downtime; tháng trung bình 730,5h x 0,5% = 3,65h = 3h39m. Con số đúng theo quy ước SRE (tháng trung bình). Khớp giữa frontmatter và §1 #1.

Kiểm §3 đo lường: công thức SLI rate 5xx/total cụ thể, burn-rate alert đa cửa sổ (1h nhanh + 6h chậm), synthetic probe bổ trợ. Đo cụ thể.

Kiểm §4 verification: đo SLI production hằng tháng, chaos/degradation test (tắt phụ thuộc phụ -> biểu đồ suy biến nhẹ), deploy test quanh cửa sổ rolling. Mỗi clause khóa có test.

Kiểm §5: burn nhanh sev-2, cạn ngân sách -> freeze, tin SLI không tin ping, chuyển rolling deploy khi bảo trì đốt ngân sách. Hợp lý.

Kiểm typo: prose ASCII thuần, tiếng Việt đủ dấu, không từ cấm. Không sửa gì.

## §3 - Kết luận

SLO định lượng với error budget tính đúng, SLI đo đúng (success-rate), chính sách burn-rate + degradation rõ, nối FR-INFRA-004 (đo). Không tìm thấy defect cần sửa. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập NFR-INFRA-002.*
