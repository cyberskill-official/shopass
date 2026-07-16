# task Status Reference - SănDeal

Nguồn sự thật cho câu hỏi "một task có thể ở trạng thái nào?". Mọi file khác (BACKLOG.md, README.md, frontmatter task) phải theo danh sách dưới đây. Phỏng theo `STATUS-REFERENCE.md` của CyberOS.

## 1. Enum đầy đủ (10 trạng thái)

Một task mang đúng một status tại mỗi thời điểm. Tất cả lowercase snake_case, trên một trục vòng đời tuyến tính.

### 1.1 Vòng đời (theo thứ tự)

| # | Status | Nghĩa | Người ghi mặc định |
|---|---|---|---|
| 1 | `draft` | Tác giả mới bắt đầu viết spec; chưa audit. | tác giả task |
| 2 | `ready_to_implement` | Spec qua audit 10/10; đủ điều kiện vào hàng đợi build. Cũng là trạng thái task quay về khi một bước trong chu trình (implementing/reviewing/testing) fail hoặc bị block (xem §1.3). | audit task / rework |
| 3 | `implementing` | Đang build; code đang được viết, test một phần đã có. | workflow ship (vào bước) |
| 4 | `ready_to_review` | Người triển khai viết xong code + test; chờ reviewer. | workflow ship |
| 5 | `reviewing` | Reviewer đang đọc diff đối chiếu mệnh đề §1 + ma trận AC §4. | workflow ship |
| 6 | `ready_to_test` | Reviewer duyệt; chờ tester. | workflow ship |
| 7 | `testing` | Tester chạy coverage gate (mỗi mệnh đề §1 có test pass trong báo cáo). | workflow ship |
| 8 | `done` | Tester chứng nhận - mọi mệnh đề truy được tới test pass; task đã ship. Terminal thành công. | workflow ship |

### 1.2 Off-ramp (do người vận hành quyết, không sức ép thời gian)

| # | Status | Nghĩa |
|---|---|---|
| 9 | `on_hold` | Hoãn có chủ đích - ngoài phạm vi wave hiện tại, sẽ xem lại sau. Vẫn ở BACKLOG như ứng viên tương lai. Hàng đợi ship mặc định bỏ qua. |
| 10 | `closed` | Kill terminal - không build (bị từ chối, bị thay bởi task khác, trùng lặp, won't-do). Vẫn ở BACKLOG để lưu vết. |

### 1.3 Khi một bước fail hoặc bị block

`[FAILED: ...]` và `[BLOCKED: ...]` KHÔNG phải là trạng thái - chúng là quyết định định tuyến:
- Fail circuit-breaker khi `implementing` (ví dụ 5 lần test fail liên tiếp trong một task) -> status rớt về `ready_to_implement`.
- Blocker không chí mạng phát hiện khi `reviewing`/`testing` (spec mơ hồ, thiếu phụ thuộc) -> status rớt về `ready_to_implement`.
- Lý do ghi vào một dòng comment trên BACKLOG hoặc một issue ngoài.

### 1.4 HITL - con người trong vòng lặp là TÙY CHỌN

Workflow ship tự lật ô status theo đường §1.1 khi mỗi cổng pass. Người vận hành có thể override bất kỳ ô nào sang ô khác bất kỳ lúc nào - không có ràng buộc chuyển trạng thái cứng. Đường workflow mặc định là gợi ý lịch sự; ô trên BACKLOG là nguồn sự thật.

Các thao tác HITL thường gặp:
- Re-audit một task đã done: lật `done` -> `ready_to_review` để buộc chạy lại cổng review + test.
- Skip review cho task tầm thường: lật `ready_to_review` -> `ready_to_test`.
- Park một task đang dở: lật `implementing` -> `on_hold`.
- Hồi sinh một task đã closed: lật `closed` -> `ready_to_implement`.

## 2. Trạng thái khởi tạo

Mọi task trong backlog SănDeal hiện ở `ready_to_implement` (đã qua audit 10/10, sẵn sàng vào hàng đợi build). Khi một agent bắt đầu build một task, lật sang `implementing` rồi theo đường §1.1 tới `done`.

## 3. Cách một agent ship chọn task tiếp theo

1. Mở [`IMPLEMENTATION-ORDER.md`](IMPLEMENTATION-ORDER.md): lấy task ở layer thấp nhất mà mọi `depends_on` đã `done`.
2. Trong cùng layer, ưu tiên `MUST` trước `SHOULD` trước `COULD`.
3. Lật status task sang `implementing`, build theo `new_files`/`sub_tasks`, chạy test §5, đối chiếu AC §4.
4. Đi tiếp theo vòng đời §1.1 tới `done`; cập nhật ô status tương ứng trên BACKLOG.

---

*Hết STATUS-REFERENCE.md (SănDeal).*
