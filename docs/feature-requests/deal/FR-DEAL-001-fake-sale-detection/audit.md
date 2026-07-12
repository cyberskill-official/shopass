---
fr_id: FR-DEAL-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập, suy lại từ file hiện tại (không tin audit co-author cũ). FR-DEAL-001 đặc tả bộ phát hiện sale ảo bằng thống kê, hiện thực đúng pseudo-code §3.5(1) của tài liệu nền từng nhánh: nạp 90 ngày `price_snapshot` qua `QueryRange`, dưới 14 điểm trả `UNKNOWN`, tính `median90`/`p10`/`trailing_min`, ba ngưỡng `1.15 / 0.97 / 1.02`, bốn verdict đóng `SALE_AO / SALE_XIN / TAM_DUOC / UNKNOWN`. Mọi phép so ngưỡng dùng integer-safe math trên BIGINT VND (nhân chéo `*100` / `*97` / `*115` / `*102`), đồng bộ DEC-PRICE-05. 13 mệnh đề §1 có AC §4 và test §5 đối ứng. Đạt 10/10.

## §2 - Findings (đã kiểm trong lượt này)

### ISS-001 - Fidelity thuật toán so §3.5(1) (đã xác nhận khớp)
Đối chiếu từng dòng pseudo-code nguồn với §1: median90=percentile50, p10=percentile10, trailing_min=min, inflated=`list>median90*1.15`, not_real=`current>=median90*0.97`, SALE_XIN=`current<=p10 AND current<=trailing_min*1.02`, dưới 14 ngày UNKNOWN. Tất cả tái tạo đúng, không thêm bớt nhánh. Không cần sửa.

### ISS-002 - Integer-safe math trên tiền tệ (đã xác nhận)
§1 #11 + `detect.go` dùng nhân chéo số nguyên thay cho nhân float (`*1.15` -> `*115` so `*100`). Khép rủi ro sai số float đúng tại biên phần trăm; test `TestDetect_BoundaryThresholds` đóng các biên `list_bang_115`, `current_bang_97`, `current_duoi_97`. Khớp DEC-PRICE-05 của FR-PRICE-002.

### ISS-003 - Typography code-comment (đã sửa)
9 mũi tên unicode `U+2192` nằm trong khối code `detect_test.go` của §5. Theo rubric O chúng được miễn (trong code fence), nhưng để thống nhất quy ước ASCII toàn repo (cùng chuẩn module SCRAPE), đã chuẩn hóa cả 9 sang `->`. Prose vốn đã sạch (không em-dash, en-dash, curly quote, ellipsis, emoji).

### ISS-004 - Hàm thuần, không sửa slice đầu vào (đã xác nhận)
`percentile` copy trước khi sort; §1 #12 cam kết tất định. Test `TestPercentile_NearestRank` kiểm slice gốc không đổi. Không cần sửa.

## §3 - Traceability §1 -> AC -> artefact (dựng từ file hiện tại)

| §1 clause | §4 AC | Test / artefact §5 / §3 |
|---|---|---|
| #1 nạp 90d QueryRange | AC #1 | service test mock repo |
| #2 <14 -> UNKNOWN | AC #2 | `TestDetect_ColdStart_Unknown` |
| #3 median90 | AC #3 | `TestPercentile_NearestRank` |
| #4 p10 | AC #4 | `TestPercentile_NearestRank` |
| #5 trailing_min | AC #5 | `TestDetect_GenuineLow_SaleXin` |
| #6 inflated 1.15 | AC #6 | `.../list_bang_115` |
| #7 not_real_discount 0.97 | AC #7 | `.../current_bang_97`, `.../current_duoi_97` |
| #8 SALE_AO | AC #8 | `TestDetect_InflatedListPrice_SaleAo` |
| #9 SALE_XIN 1.02 | AC #9 | `TestDetect_SaleXin_BoundaryAtFloor` |
| #10 TAM_DUOC | AC #10 | `TestDetect_Middle_TamDuoc` |
| #11 integer-safe math | AC #11 | `TestDetect_BoundaryThresholds` |
| #12 thuần/tất định | AC #12 | `TestPercentile_NearestRank` (slice gốc) |
| #13 enum đóng Verdict | AC #13 | `types.go` 4 hằng |

## §4 - Kết luận

Mỗi mệnh đề normative ánh xạ 1:1 với một dòng §3.5(1) và có test backing; không mệnh đề mồ côi. Ba hằng ngưỡng load-bearing tái tạo đúng; integer-safe math khép rủi ro float; cold-start trung thực bàn giao FR-DEAL-002. Đã sửa surgical 9 mũi tên comment sang ASCII. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập FR-DEAL-001.*
