---
id: FR-CART-006
title: "Checklist xu/coin - nhắc nhở nhiệm vụ xu hằng ngày (đọc trạng thái hiển thị + hiện checklist), KHÔNG auto-click (tự động hóa xu rủi ro ban High §3.9c); user tự bấm trên sàn"
module: CART
priority: SHOULD
status: ready_to_implement
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-EXT-002, FR-CART-005, FR-NOTIF-001, NFR-AFFIL-001]
depends_on: [FR-EXT-002]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.9 (xếp hạng rủi ro ban: tự động hóa xu/voucher rủi ro High -> SănDeal chỉ làm checklist nhắc nhở + auto-test mã user-initiated, KHÔNG tự động click xu)"
  - "docs/... §6 (catalog tính năng: checklist xu), §1.2 (persona Huy: săn xu, cần checklist xu)"
source_decisions:
  - "DEC-CART-30: checklist xu CHỈ nhắc nhở (đọc trạng thái nhiệm vụ hiển thị + hiện danh sách việc cần làm) - TUYỆT ĐỐI KHÔNG auto-click/tự hoàn thành nhiệm vụ xu (§3.9c xếp tự động hóa xu rủi ro ban High)"
  - "DEC-CART-31: chỉ ĐỌC trạng thái nhiệm vụ xu hiển thị (đã làm/chưa làm) qua khung reader FR-EXT-002; KHÔNG mô phỏng click, KHÔNG gọi API hoàn thành nhiệm vụ"
  - "DEC-CART-32: checklist là dữ liệu nhắc (task_type, due_date, done) - hiện cho user tự bấm trên sàn; mọi hành động hoàn thành do user tự làm"
  - "DEC-CART-33: dùng user_activity_coin_task (§3.4) lưu trạng thái nhắc per user/platform; reminder qua FR-NOTIF-001 (nhắc nhẹ, không spam)"
  - "DEC-CART-34: KHÔNG thu thập credential khi đọc trạng thái xu (đồng nhất ranh giới FR-EXT-002/003); chỉ trạng thái done/chưa-done của nhiệm vụ hiển thị"

language: "TypeScript 5.x (extension reader, chỉ-đọc) + Go 1.22 (cart-svc lưu trạng thái checklist); KHÔNG tự động hóa click"
service: shopass/services/cart/
new_files:
  - extension/src/content/shared/coin-task-reader.ts
  - extension/src/ui/coin-checklist.ts
  - db/migrations/0012_coin_task.up.sql
  - db/migrations/0012_coin_task.down.sql
  - services/cart/internal/coin/types.go
  - services/cart/internal/coin/repo.go
  - extension/test/coin-no-autoclick.test.ts
  - services/cart/internal/coin/repo_test.go
modified_files:
  - extension/src/shared/types.ts                # thêm CoinTask, CoinChecklistMessage
allowed_tools:
  - file_read: extension/**
  - file_read: services/cart/**
  - file_read: db/migrations/**
  - file_write: extension/**
  - file_write: services/cart/**
  - file_write: db/migrations/**
  - bash: cd extension && npm test; cd services/cart && go test ./...
disallowed_tools:
  - auto-click / tự hoàn thành nhiệm vụ xu (vi phạm DEC-CART-30, §3.9c rủi ro ban High)
  - mô phỏng click / gọi API hoàn thành nhiệm vụ xu (vi phạm DEC-CART-31)
  - thu thập credential khi đọc trạng thái xu (vi phạm DEC-CART-34)
  - spam reminder (vi phạm DEC-CART-33, nhắc nhẹ)

effort_hours: 5
sub_tasks:
  - "0.5h: types.ts - CoinTask{taskType, dueDate, done}, CoinChecklistMessage (chỉ trạng thái hiển thị)"
  - "1.0h: coin-task-reader.ts - ĐỌC trạng thái nhiệm vụ xu hiển thị (done/chưa); KHÔNG click, KHÔNG gọi API hoàn thành"
  - "0.75h: coin-checklist.ts - UI hiện checklist + nút dẫn user tới trang nhiệm vụ trên sàn (user tự bấm)"
  - "0.5h: 0012_coin_task.up/down.sql - bảng user_activity_coin_task (user_id, platform_id, task_type, due_date, done)"
  - "0.75h: coin/types.go + repo.go - lưu/đọc trạng thái checklist per user/platform"
  - "1.0h: coin-no-autoclick.test.ts - khẳng định KHÔNG click/dispatch event/gọi API hoàn thành nhiệm vụ"
  - "0.5h: repo_test.go - upsert trạng thái task, đọc theo user/platform, scope user"

risk_if_skipped: "Checklist xu là tiện ích cho persona Huy (săn xu, §1.2) - nhắc các nhiệm vụ xu hằng ngày để không bỏ lỡ. Là SHOULD nên không chặn release. Nhưng đây là FR mà ranh giới phải tuyệt đối đúng: §3.9c xếp tự động hóa xu/voucher ở mức rủi ro ban CAO NHẤT (High) - cao hơn cả scraping - vì sàn coi tự động hoàn thành nhiệm vụ xu là abuse/farming trực tiếp. Tài liệu nêu rõ giải pháp của SănDeal là CHỈ checklist nhắc nhở, KHÔNG tự động click xu. Nếu làm SAI - auto-click hoàn thành nhiệm vụ, hoặc mô phỏng click - thì kích hoạt đúng cơ chế ban High, làm tài khoản người dùng bị khóa và SănDeal bị coi là công cụ farming (đụng cả gian lận user §5.3 và niềm tin §5.4). Ranh giới: chỉ ĐỌC trạng thái nhiệm vụ hiển thị và NHẮC; mọi hành động hoàn thành do user tự bấm trên sàn. Phải test khẳng định không có đường auto-click, không chỉ quy ước. Nếu thu thập credential khi đọc trạng thái thì phá cam kết tối thiểu hóa."
---

## §1 - Mô tả (BCP-14 normative)

Extension và service CART **MUST** cung cấp checklist xu CHỈ nhắc nhở - đọc trạng thái nhiệm vụ xu hiển thị và hiện danh sách việc cần làm cho người dùng tự bấm trên sàn - và TUYỆT ĐỐI không tự động click/hoàn thành nhiệm vụ xu. Hợp đồng:

1. Checklist xu **MUST** chỉ nhắc nhở (DEC-CART-30): đọc trạng thái nhiệm vụ xu hiển thị (đã làm/chưa làm) và hiện danh sách việc; TUYỆT ĐỐI KHÔNG auto-click, KHÔNG tự hoàn thành nhiệm vụ xu (§3.9c xếp tự động hóa xu rủi ro ban High).
2. `coin-task-reader.ts` **MUST** chỉ ĐỌC trạng thái nhiệm vụ xu hiển thị qua khung reader FR-EXT-002 (DEC-CART-31): KHÔNG mô phỏng click (`element.click()`, dispatch sự kiện chuột), KHÔNG gọi API hoàn thành nhiệm vụ.
3. Checklist **MUST** là dữ liệu nhắc (DEC-CART-32): `task_type` (loại nhiệm vụ), `due_date` (hạn), `done` (đã làm chưa) - hiển thị cho người dùng; nút trên UI chỉ DẪN người dùng tới trang nhiệm vụ trên sàn để tự bấm.
4. Service **MUST** lưu trạng thái nhắc qua bảng `user_activity_coin_task` (§3.4): `(user_id, platform_id, task_type, due_date, done)` - per user/platform (DEC-CART-33).
5. Reminder **MUST** gửi qua FR-NOTIF-001 dạng nhắc nhẹ (DEC-CART-33), KHÔNG spam - tối đa nhắc hợp lý mỗi ngày, tôn trọng tùy chọn người dùng.
6. Quá trình đọc trạng thái xu **MUST NOT** thu thập credential (DEC-CART-34): không đọc cookie/token (đồng nhất ranh giới FR-EXT-002/003); chỉ lấy trạng thái done/chưa-done của nhiệm vụ hiển thị.
7. UI checklist **MUST** làm rõ rằng SănDeal chỉ nhắc, người dùng tự thực hiện trên sàn (minh bạch, hợp triết lý hậu-Honey) - không tạo ấn tượng tự động hoàn thành.
8. Service **MUST** scope trạng thái theo `user_id`: người dùng chỉ đọc checklist của chính mình (chống truy cập chéo).
9. Đọc trạng thái **MUST** xử lý lịch sự khi không lấy được (DOM đổi, chưa đăng nhập): hiện "chưa lấy được trạng thái", không lỗi vỡ; KHÔNG thử click để "kích hoạt".
10. Toàn bộ đường đọc **MUST** local-first (FR-TRUST-002): trạng thái đọc trên máy client; chỉ trạng thái tối thiểu (done/chưa) đồng bộ nếu cần nhắc.
11. Là tính năng SHOULD: có thể hoãn sau MVP; khi làm **MUST** giữ ranh giới chỉ-nhắc tuyệt đối.
12. `npm test` (extension) + `go test ./...` (cart-svc) xanh; `tsc --noEmit` sạch.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao CHỈ nhắc, không auto-click (DEC-CART-30)?** §3.9c xếp tự động hóa xu/voucher ở mức rủi ro ban CAO NHẤT (High) - cao hơn cả scraping - vì sàn coi tự động hoàn thành nhiệm vụ xu là farming trực tiếp. Tài liệu nêu rõ giải pháp của SănDeal là checklist nhắc nhở, KHÔNG tự động click. Auto-click kích hoạt đúng cơ chế ban này, khóa tài khoản người dùng và biến SănDeal thành công cụ farming - đụng cả gian lận user (§5.3) và niềm tin (§5.4). Chỉ nhắc là phía an toàn tuyệt đối.

**Vì sao chỉ đọc trạng thái hiển thị (DEC-CART-31)?** Ranh giới giữa "trợ lý nhắc" và "bot farming" là ai bấm nút. Đọc trạng thái nhiệm vụ (đã làm/chưa) để nhắc là chỉ-đọc, an toàn. Mô phỏng click hay gọi API hoàn thành là thay người dùng thực hiện - vượt ranh giới. Reader chỉ đọc, không bao giờ click.

**Vì sao user tự bấm trên sàn (DEC-CART-32)?** SănDeal hiện checklist "hôm nay bạn chưa điểm danh, chưa xem live" và nút dẫn tới trang nhiệm vụ; người dùng tự bấm điểm danh trên sàn. Quyền và hành động ở người dùng - vừa tránh ban, vừa minh bạch (không giả vờ tự động giúp). Đây là cùng triết lý với auto-test mã user-initiated (FR-CART-005).

**Vì sao không spam reminder (DEC-CART-33)?** Nhắc quá nhiều phản tác dụng - người dùng tắt thông báo. Nhắc nhẹ (một lần mỗi ngày cho nhiệm vụ chưa làm, tôn trọng tùy chọn) giữ giá trị mà không phiền. Reminder qua FR-NOTIF-001 với tần suất hợp lý.

**Vì sao không thu thập credential (DEC-CART-34)?** Đồng nhất ranh giới niềm tin lõi (FR-EXT-002/003): đọc trạng thái nhiệm vụ hiển thị không cần và không được chạm cookie/token. Chỉ lấy done/chưa-done. Giữ cam kết tối thiểu hóa ngay cả ở tính năng phụ này.

**Vì sao là SHOULD nhưng ranh giới tuyệt đối (§1 #11)?** Checklist xu tiện nhưng không chặn release - có thể hoãn. Tuy nhiên khi làm, ranh giới chỉ-nhắc không được nới: một lần auto-click là rủi ro ban High cho người dùng. SHOULD về ưu tiên, MUST về ranh giới an toàn.

---

## §3 - Hợp đồng API / DDL

### Reader trạng thái (coin-task-reader.ts) - CHỈ đọc, KHÔNG click

```ts
// extension/src/content/shared/coin-task-reader.ts
export interface CoinTask {
  taskType: string;   // 'daily_checkin' | 'watch_live' | ... (loại nhiệm vụ hiển thị)
  done: boolean;      // đã làm chưa (đọc từ trạng thái hiển thị)
}

// readCoinTasks CHỈ đọc trạng thái nhiệm vụ xu hiển thị (DEC-CART-31).
// TUYỆT ĐỐI KHÔNG click, KHÔNG dispatch sự kiện, KHÔNG gọi API hoàn thành.
export function readCoinTasks(): CoinTask[] {
  const rows = queryCoinTaskRows(); // đọc DOM khu nhiệm vụ xu (khung reader FR-EXT-002)
  return rows.map((el) => ({
    taskType: parseTaskType(el),
    done: parseDoneState(el),       // chỉ đọc trạng thái done/chưa
  }));
  // KHÔNG có el.click(), KHÔNG dispatchEvent, KHÔNG fetch hoàn thành nhiệm vụ
}
```

### Migration (golang-migrate)

```sql
-- db/migrations/0012_coin_task.up.sql
CREATE TABLE user_activity_coin_task (
  id          BIGSERIAL   PRIMARY KEY,
  user_id     BIGINT      NOT NULL REFERENCES app_user(id),
  platform_id SMALLINT    NOT NULL REFERENCES platform(id),
  task_type   TEXT        NOT NULL,
  due_date    DATE        NOT NULL,
  done        BOOLEAN     NOT NULL DEFAULT false,
  UNIQUE (user_id, platform_id, task_type, due_date)
);
CREATE INDEX idx_coin_due ON user_activity_coin_task (user_id, due_date) WHERE done = false;

-- db/migrations/0012_coin_task.down.sql
DROP TABLE user_activity_coin_task;
```

### Repo (repo.go) - scope user, lưu trạng thái nhắc

```go
// services/cart/internal/coin/repo.go
func (r *Repo) UpsertTask(ctx context.Context, userID int64, platformID int16, t CoinTask) error {
    _, err := r.pool.Exec(ctx,
        `INSERT INTO user_activity_coin_task (user_id, platform_id, task_type, due_date, done)
         VALUES ($1,$2,$3,$4,$5)
         ON CONFLICT (user_id, platform_id, task_type, due_date) DO UPDATE SET done = EXCLUDED.done`,
        userID, platformID, t.TaskType, t.DueDate, t.Done)
    return err
}

// ListPending trả nhiệm vụ chưa làm của ĐÚNG user (scope user_id, DEC-CART... §1 #8).
func (r *Repo) ListPending(ctx context.Context, userID int64, day time.Time) ([]CoinTask, error) {
    // ... WHERE user_id = $1 AND due_date = $2 AND done = false
}
```

### UI checklist (coin-checklist.ts) - dẫn user tự bấm

```ts
// extension/src/ui/coin-checklist.ts
// Hiện checklist + nút DẪN tới trang nhiệm vụ trên sàn; user TỰ bấm (DEC-CART-32).
function renderChecklist(tasks: CoinTask[]) {
  for (const t of tasks) {
    const row = makeRow(t.taskType, t.done);
    if (!t.done) {
      row.appendLink("Tới trang nhiệm vụ", taskUrlOnPlatform(t.taskType)); // mở để user tự làm
      // KHÔNG có nút "tự động hoàn thành"
    }
  }
  showNote("SănDeal chỉ nhắc; bạn tự thực hiện trên sàn."); // minh bạch (§1 #7)
}
```

---

## §4 - Acceptance criteria

1. Grep `extension/src/content/shared/coin-task-reader.ts` + `coin-checklist.ts`: KHÔNG có `.click(`, KHÔNG `dispatchEvent`, KHÔNG `fetch`/API gọi hoàn thành nhiệm vụ xu.
2. `readCoinTasks` trả danh sách `CoinTask{taskType, done}` đọc từ trạng thái hiển thị (không thay đổi trạng thái).
3. UI checklist hiện nhiệm vụ chưa làm kèm link DẪN tới trang nhiệm vụ trên sàn; KHÔNG có nút "tự động hoàn thành".
4. UI có ghi chú minh bạch "SănDeal chỉ nhắc; bạn tự thực hiện".
5. Migration tạo `user_activity_coin_task` với `UNIQUE(user_id, platform_id, task_type, due_date)`.
6. `UpsertTask` lưu/cập nhật trạng thái; `ListPending` trả nhiệm vụ chưa làm của đúng user (scope user_id).
7. Reader KHÔNG thu thập credential: grep KHÔNG `document.cookie`/token khi đọc trạng thái xu.
8. Reminder gửi qua FR-NOTIF-001 nhắc nhẹ (không spam); tôn trọng tùy chọn.
9. Đọc thất bại (DOM đổi/chưa đăng nhập) -> "chưa lấy được trạng thái", không lỗi, KHÔNG thử click.
10. Trạng thái scope user_id: user A không đọc checklist user B.
11. `npm test` + `go test ./...` xanh; `tsc --noEmit` sạch.

---

## §5 - Kiểm thử (verification)

```ts
// extension/test/coin-no-autoclick.test.ts
import { readFile } from "fs/promises";

test("reader + UI KHÔNG auto-click / dispatch / gọi API hoàn thành", async () => {
  for (const f of ["src/content/shared/coin-task-reader.ts", "src/ui/coin-checklist.ts"]) {
    const src = await readFile(f, "utf8");
    expect(src).not.toMatch(/\.click\s*\(/);        // không mô phỏng click
    expect(src).not.toMatch(/dispatchEvent/);       // không phát sự kiện chuột
    expect(src).not.toMatch(/complete.?task|claim.?coin|hoàn thành nhiệm vụ/i); // không gọi hoàn thành
  }
});

test("reader chỉ đọc trạng thái, KHÔNG chạm credential", async () => {
  const src = await readFile("src/content/shared/coin-task-reader.ts", "utf8");
  expect(src).not.toMatch(/document\.cookie/);
  expect(src).not.toMatch(/Authorization|token/i);
});

test("readCoinTasks trả trạng thái done không đổi DOM", () => {
  document.body.innerHTML = coinTaskFixture; // 2 nhiệm vụ: 1 done, 1 chưa
  const before = document.body.innerHTML;
  const tasks = readCoinTasks();
  expect(tasks).toHaveLength(2);
  expect(tasks.some((t) => t.done)).toBe(true);
  expect(document.body.innerHTML).toBe(before); // đọc không sửa DOM (không click)
});
```

```go
// services/cart/internal/coin/repo_test.go
func TestUpsertTask_Idempotent(t *testing.T) {
    r, uid := setupCoin(t)
    task := CoinTask{TaskType: "daily_checkin", DueDate: today, Done: false}
    require.NoError(t, r.UpsertTask(ctx, uid, 1, task))
    task.Done = true
    require.NoError(t, r.UpsertTask(ctx, uid, 1, task)) // cùng khóa -> update done
    pending, _ := r.ListPending(ctx, uid, today)
    require.Empty(t, pending) // đã done -> không còn pending
}

func TestListPending_ScopedToUser(t *testing.T) {
    r, uidA := setupCoin(t)
    r.UpsertTask(ctx, uidA, 1, CoinTask{TaskType: "watch_live", DueDate: today, Done: false})
    uidB := makeUser(t, r)
    pendingB, _ := r.ListPending(ctx, uidB, today)
    require.Empty(t, pendingB) // user B không thấy nhiệm vụ của A
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `types.ts` (CoinTask, CoinChecklistMessage) -> `coin-task-reader.ts` (CHỈ đọc trạng thái hiển thị, không click/API) -> `0012_coin_task.up/down.sql` (bảng user_activity_coin_task) -> `coin/types.go` + `repo.go` (lưu/đọc scope user) -> `coin-checklist.ts` (UI hiện checklist + link dẫn user tự bấm, ghi chú minh bạch) -> tests. Migration 0012 nối tiếp 0011 (FR-CART-002). Reader tái dùng khung FR-EXT-002 nhưng tuyệt đối chỉ-đọc (không có ngoại lệ áp-gỡ như FR-CART-005 vì xu không có thao tác đo). Reminder nối FR-NOTIF-001 nhắc nhẹ. Test then chốt là grep khẳng định không auto-click/dispatch/API-hoàn-thành - ranh giới chống ban High phải kiểm chứng được.

---

## §7 - Phụ thuộc

- **FR-EXT-002** - khung content script + reader (chỉ-đọc); `coin-task-reader` đọc trạng thái nhiệm vụ xu hiển thị, tuyệt đối không click (depends_on cứng).
- **FR-NOTIF-001 (downstream)** - template + routing để gửi reminder nhắc nhẹ; checklist không tự gửi, dùng kênh notif.
- **FR-CART-005 (sibling)** - auto-test mã cùng triết lý user-initiated/không-tự-động; checklist xu còn nghiêm hơn (không có cả thao tác đo, chỉ nhắc).
- **NFR-AFFIL-001** - guardrails compliance (né Honey, Chrome policy); checklist chỉ-nhắc là một mặt.
- Extension/lib: TypeScript (reader + UI), Go (repo); golang-migrate; test Jest + jsdom, Go test.

---

## §8 - Payload ví dụ

### Checklist hiển thị (nhắc, user tự bấm trên sàn)

```json
{
  "platform": "shopee",
  "date": "2026-06-28",
  "tasks": [
    { "taskType": "daily_checkin", "done": false },
    { "taskType": "watch_live",    "done": true  },
    { "taskType": "browse_shops",  "done": false }
  ]
}
```

### Luồng (mô tả, không phải payload)

```
extension đọc trạng thái nhiệm vụ xu hiển thị (CHỈ đọc)
  -> hiện checklist: việc nào chưa làm
  -> với việc chưa làm: link "Tới trang nhiệm vụ" (user TỰ bấm trên sàn)
  -> reminder nhắc nhẹ qua FR-NOTIF-001 (không spam)
  -> KHÔNG auto-click, KHÔNG tự hoàn thành (chống ban High §3.9c)
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Đồng bộ trạng thái checklist giữa thiết bị - thêm khi cần; giữ scope user.
- Nhắc thông minh theo thói quen người dùng (giờ hay quên điểm danh) - tối ưu reminder sau, không đổi ranh giới chỉ-nhắc.
- Mở rộng đọc nhiệm vụ xu TikTok Shop/Lazada (khung FR-EXT-007/008) - per-sàn sau, vẫn chỉ-đọc.
- Gamification điểm SănDeal nội bộ cho việc hoàn thành (§6 #10) - tách FR riêng, không auto-click sàn.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Auto-click hoàn thành nhiệm vụ | coin-no-autoclick test + grep | rủi ro ban High, khóa tài khoản | CHỈ nhắc, không click (DEC-CART-30) |
| Mô phỏng click / dispatch event | grep .click/dispatchEvent | farming -> ban | Chỉ đọc trạng thái (DEC-CART-31) |
| Gọi API hoàn thành nhiệm vụ | grep complete/claim | né disclosure, abuse | Không API hoàn thành (DEC-CART-31) |
| Thu thập credential khi đọc xu | grep cookie/token | phá tối thiểu hóa | Chỉ done/chưa-done (DEC-CART-34) |
| Spam reminder | review tần suất | user tắt thông báo | Nhắc nhẹ qua NOTIF-001 (DEC-CART-33) |
| Đọc thất bại thử click kích hoạt | fail-state test | vượt ranh giới | "Chưa lấy được trạng thái", không click (§1 #9) |
| Đọc checklist user khác | scope test | rò rỉ | Scope user_id (§1 #8) |
| UI tạo ấn tượng tự động | review note | hiểu lầm + lệch triết lý | Ghi chú "bạn tự thực hiện" (§1 #7) |
| Sửa DOM khi đọc | no-mutate test | tác động ngoài ý muốn | Đọc không sửa DOM (§5) |

---

## §11 - Ghi chú

- Checklist xu là tiện ích cho persona Huy (SHOULD); ranh giới chỉ-nhắc là MUST tuyệt đối.
- §3.9c xếp tự động hóa xu rủi ro ban CAO NHẤT (High) - cao hơn scraping; giải pháp SănDeal là checklist nhắc, KHÔNG auto-click.
- Ranh giới giữa trợ lý nhắc và bot farming là ai bấm nút: SănDeal đọc trạng thái + nhắc; người dùng tự bấm trên sàn.
- Không thu thập credential khi đọc trạng thái - đồng nhất cam kết tối thiểu hóa lõi (FR-EXT-002/003).
- Reminder nhắc nhẹ qua FR-NOTIF-001, không spam; tôn trọng tùy chọn người dùng.
- Test then chốt là grep khẳng định không auto-click/dispatch/API-hoàn-thành - ranh giới chống ban High phải kiểm chứng được, không chỉ quy ước.

---

*Hết FR-CART-006. Status: ready_to_implement (mục tiêu audit 10/10).*
