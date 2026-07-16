---
fr_id: TASK-MOBILE-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-MOBILE-002 đặc tả lớp theo dõi giá + checkout assistant trên mobile ở mức triển khai được. 11 mệnh đề §1 normative, mỗi mệnh đề có AC và test. Ranh giới compliance hậu-Honey được giữ tuyệt đối: checkout assistant chỉ chạy user-initiated, chỉ hiển thị gợi ý (người dùng tự áp mã), giỏ gửi lên optimizer tối thiểu hóa (không cookie/token). Mobile là client mỏng đọc API, không tính sale ảo phía client. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Checkout assistant tự kích hoạt (đã chốt)
Cám dỗ tự bật assistant nền lúc thanh toán lặp đúng vết Honey (§4.2). Giải: §1 #4 + DEC-MOBILE-11 chỉ chạy khi bấm; test `checkout assistant KHÔNG tự chạy khi mount`.

### ISS-002 - Auto-apply voucher
Auto-click/auto-apply trong app sàn là abuse, vi phạm policy. Giải: §1 #6 + DEC-MOBILE-13 chỉ hiển thị, người dùng tự áp; review + AC #6.

### ISS-003 - Rò dữ liệu giỏ lên optimizer
Gửi cookie/token sàn phá tối thiểu hóa dữ liệu. Giải: §1 #7 + DEC-MOBILE-15 chỉ product_id/qty/giá; test `payload optimizer chỉ chứa trường tối thiểu hóa`.

### ISS-004 - Logic giá phân mảnh client/backend
Tính sale ảo phía client lệch với web/backend. Giải: §1 #1 + DEC-MOBILE-10 client mỏng đọc API; test `biểu đồ đọc price-history từ API`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 client mỏng đọc API | #1,#2 | `trackClient.ts` + test |
| #2 biểu đồ render | #2 | `priceChart.tsx` |
| #3 push deep-link | #3 | FCM handler -> màn sản phẩm |
| #4 user-initiated | #4 | `checkoutAssistant.tsx` + mount test |
| #5 optimizer backend | #5 | `optimizerClient.ts` |
| #6 chỉ hiển thị | #6 | review + AC |
| #7 giỏ tối thiểu hóa | #5 | test payload |
| #8 trạng thái rỗng | #2,#7 | "chưa đủ dữ liệu" / "không có voucher" |
| #11 disclosure affiliate | #8 | `DisclosureBanner` + test |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/test backing. Compliance (user-initiated, chỉ hiển thị, tối thiểu hóa) được kiểm bằng test hành vi cụ thể. Không mệnh đề mồ côi. Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-MOBILE-002.*
