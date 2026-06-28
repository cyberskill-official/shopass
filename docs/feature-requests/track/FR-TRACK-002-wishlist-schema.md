---
id: FR-TRACK-002
title: "Schema + API wishlist / wishlist_item - danh sách mong muốn nhiều cấp với target_price BIGINT VND, CRUD có phân quyền theo chủ sở hữu, FK tới tracked_product"
module: TRACK
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-TRACK-001, FR-TRACK-003, FR-WEB-004, FR-INFRA-002]
depends_on: [FR-TRACK-001]
blocks: [FR-WEB-004]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model wishlist + wishlist_item target_price)"
  - "docs/... §3.7 (API quản lý wishlist cho web app)"
source_decisions:
  - "DEC-TRACK-10: wishlist là nhóm do user đặt tên (id, user_id, name); một user có nhiều wishlist (vd 'Cần mua', 'Theo dõi sau')"
  - "DEC-TRACK-11: wishlist_item gắn một tracked_product vào một wishlist kèm target_price (BIGINT VND, nullable - giá mong muốn)"
  - "DEC-TRACK-12: mọi thao tác CRUD bắt buộc kiểm chủ sở hữu - user chỉ đọc/sửa/xóa wishlist của chính mình (chống IDOR truy cập chéo user)"
  - "DEC-TRACK-13: thêm item vào wishlist yêu cầu product_id đã tồn tại trong tracked_product (đã qua FR-TRACK-001); UNIQUE(wishlist_id, product_id) chống thêm trùng"
  - "DEC-TRACK-14: target_price lưu BIGINT VND đồng nhất DEC-PRICE-05; là nguồn cho alert_rule type price_below (FR-TRACK-003) khi user đặt giá mong muốn"

language: "Go 1.22 (track-svc); PostgreSQL 16"
service: shopass/services/track/
new_files:
  - services/track/migrations/0002_wishlist.sql
  - services/track/internal/api/wishlist.go
  - services/track/internal/track/wishlist_repo.go
  - services/track/internal/track/wishlist_repo_test.go
  - services/track/internal/api/wishlist_test.go
modified_files:
  - services/track/internal/api/router.go            # đăng ký các route wishlist
allowed_tools:
  - file_read: services/track/**
  - file_write: services/track/**
  - bash: cd services/track && go test ./...
disallowed_tools:
  - bỏ kiểm chủ sở hữu trên bất kỳ route nào (vi phạm DEC-TRACK-12, lỗ IDOR đọc/sửa wishlist người khác)
  - lưu target_price dạng float/numeric thập phân (vi phạm DEC-TRACK-14, sai số tiền tệ)
  - thêm wishlist_item với product_id chưa có trong tracked_product (vi phạm DEC-TRACK-13, FK rác)

effort_hours: 5
sub_tasks:
  - "0.5h: 0002_wishlist.sql - bảng wishlist + wishlist_item + FK + UNIQUE(wishlist_id, product_id)"
  - "1.0h: wishlist_repo.go - CreateWishlist, ListWishlists(user), AddItem, RemoveItem, DeleteWishlist (đều scope user_id)"
  - "1.0h: wishlist.go - 5 handler CRUD + kiểm chủ sở hữu + 201/200/403/404/400"
  - "0.5h: router.go - đăng ký route sau JWT middleware (FR-INFRA-001)"
  - "2.0h: wishlist_repo_test.go + wishlist_test.go - 7 test (create, list scope user, add item, dup item, target_price int64, cross-user 403/404, delete cascade)"

risk_if_skipped: "wishlist là cách người dùng tổ chức các sản phẩm họ muốn mua và đặt giá mong muốn (target_price) - không có nó thì người dùng chỉ có một danh sách phẳng các SKU đã track, không nhóm được theo ý định và không khai báo được mức giá họ chờ. target_price ở đây là nguồn dữ liệu cho alert_rule type price_below (FR-TRACK-003): thiếu wishlist thì luồng 'báo tôi khi món này về giá X' mất đầu vào. Nguy hiểm nhất: nếu bỏ kiểm chủ sở hữu trên các route CRUD thì một user đoán id wishlist của người khác là đọc/sửa/xóa được (lỗ IDOR) - rò rỉ ý định mua sắm và phá dữ liệu người dùng khác, một vi phạm quyền riêng tư trực tiếp đụng PDPL (FR-COMPLY). Lưu target_price dạng float gây sai số trên so sánh giá của engine alert."
---

## §1 - Mô tả (BCP-14 normative)

Service TRACK **MUST** cung cấp schema và API CRUD cho `wishlist` (nhóm do user đặt tên) và `wishlist_item` (một SKU trong nhóm kèm `target_price`), với mọi thao tác bắt buộc kiểm chủ sở hữu. Hợp đồng:

1. **MUST** định nghĩa bảng `wishlist (id BIGSERIAL PK, user_id BIGINT REFERENCES app_user(id), name TEXT NOT NULL, created_at TIMESTAMPTZ DEFAULT now())`.
2. **MUST** định nghĩa bảng `wishlist_item (id BIGSERIAL PK, wishlist_id BIGINT REFERENCES wishlist(id) ON DELETE CASCADE, product_id BIGINT REFERENCES tracked_product(id), target_price BIGINT, added_at TIMESTAMPTZ DEFAULT now())`.
3. **MUST** lưu `target_price` dạng `BIGINT` VND nullable (DEC-TRACK-14): NULL nghĩa "chưa đặt giá mong muốn". KHÔNG dùng float/numeric (đồng nhất DEC-PRICE-05).
4. **MUST** đặt `UNIQUE (wishlist_id, product_id)` (DEC-TRACK-13): một SKU không thể nằm hai lần trong cùng một wishlist; là khóa `ON CONFLICT` cho thêm idempotent.
5. **MUST** phục vụ `POST /v1/wishlists {name}` tạo wishlist mới gắn `user_id` từ JWT; trả `201` + `{id, name}`.
6. **MUST** phục vụ `GET /v1/wishlists` trả mọi wishlist của đúng user gọi (kèm số item); KHÔNG trả wishlist của user khác (DEC-TRACK-12).
7. **MUST** phục vụ `POST /v1/wishlists/{id}/items {product_id, target_price?}` thêm một SKU; `product_id` phải tồn tại trong `tracked_product` (DEC-TRACK-13), nếu không trả `400`. Thêm trùng (SKU đã trong wishlist) trả `200` idempotent (no-op), không tạo dòng thứ hai.
8. **MUST** phục vụ `DELETE /v1/wishlists/{id}/items/{product_id}` gỡ một SKU khỏi wishlist; `DELETE /v1/wishlists/{id}` xóa cả wishlist và mọi item (CASCADE).
9. **MUST** kiểm chủ sở hữu trên MỌI route theo `{id}` (DEC-TRACK-12): nếu `wishlist.user_id != caller` thì trả `404` (không `403`) để không lộ sự tồn tại của wishlist người khác. KHÔNG có route nào bỏ qua kiểm tra này.
10. **MUST** lấy `user_id` từ JWT do API Gateway (FR-INFRA-001) gắn; handler KHÔNG tự parse token.
11. **MUST** trả `target_price` trong JSON là số nguyên VND (int64) hoặc `null`; KHÔNG float, KHÔNG string. Đặt `Content-Type: application/json; charset=utf-8`.
12. **SHOULD** phát OTel: `wishlist_ops_total{op, status}` (counter, `op - {create, list, add_item, remove_item, delete}`), `wishlist_item_count` (gauge theo user) để theo dõi mức dùng.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao wishlist nhiều cấp do user đặt tên (DEC-TRACK-10)?** Người mua sắm phân loại ý định: "cần mua tháng này", "chờ sale 12.12", "quà tặng". Một danh sách phẳng trộn mọi thứ lại. Cho user tạo nhiều wishlist có tên là mô hình quen thuộc (giống giỏ lưu của các sàn) và là khung để gắn `target_price` theo từng món trong từng ngữ cảnh.

**Vì sao kiểm chủ sở hữu trả 404 thay vì 403 (DEC-TRACK-12, §1 #9)?** Trả `403 Forbidden` cho id của người khác vô tình xác nhận "wishlist này tồn tại nhưng không phải của bạn" - rò rỉ thông tin. Trả `404 Not Found` đồng nhất cho cả "không tồn tại" lẫn "không phải của bạn" làm kẻ dò không phân biệt được, đóng kênh liệt kê id. Đây là lỗ IDOR kinh điển nếu bỏ kiểm tra: id tuần tự (BIGSERIAL) dễ đoán.

**Vì sao target_price là BIGINT VND nullable (DEC-TRACK-14)?** Giá mong muốn là số nguyên đồng. NULL phân biệt "thêm vào để theo dõi" với "thêm vào và chờ về giá X". Giá trị này chảy thẳng vào alert_rule type `price_below` (FR-TRACK-003): khi user đặt target, hệ thống có thể tạo luật cảnh báo tương ứng. BIGINT giữ phép so sánh giá chính xác tuyệt đối.

**Vì sao UNIQUE(wishlist_id, product_id) và thêm trùng là no-op (DEC-TRACK-13, §1 #7)?** Người dùng có thể bấm "thêm vào wishlist" hai lần. Hai dòng cho cùng SKU trong một danh sách là vô nghĩa và làm rối UI. Ràng buộc kép + `ON CONFLICT DO NOTHING` (hoặc `DO UPDATE` target_price) làm thao tác idempotent: bấm lại chỉ cập nhật giá mong muốn, không nhân đôi.

**Vì sao product_id phải tồn tại trong tracked_product (§1 #7)?** `wishlist_item` chỉ trỏ tới SKU hệ thống đã biết và đang theo dõi giá. Cho thêm một `product_id` tùy ý tạo FK rác trỏ tới hư không, và sau không có chuỗi giá để hiện. Buộc SKU đã qua `POST /v1/track` (FR-TRACK-001) giữ wishlist luôn gắn dữ liệu thật.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/track/migrations/0002_wishlist.sql
CREATE TABLE wishlist (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  name       TEXT        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_wishlist_user ON wishlist (user_id);

CREATE TABLE wishlist_item (
  id           BIGSERIAL   PRIMARY KEY,
  wishlist_id  BIGINT      NOT NULL REFERENCES wishlist(id) ON DELETE CASCADE,
  product_id   BIGINT      NOT NULL REFERENCES tracked_product(id),
  target_price BIGINT      CHECK (target_price IS NULL OR target_price > 0), -- VND
  added_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (wishlist_id, product_id)
);
```

### Repo có phân quyền (Go)

```go
// services/track/internal/track/wishlist_repo.go

// AddItem thêm SKU vào wishlist (idempotent), cập nhật target_price nếu thêm lại.
// product_id phải tồn tại trong tracked_product (FK chặn id rác).
func (r *Repo) AddItem(ctx context.Context, wishlistID, productID int64, target *int64) error {
    _, err := r.pool.Exec(ctx,
        `INSERT INTO wishlist_item (wishlist_id, product_id, target_price)
         VALUES ($1, $2, $3)
         ON CONFLICT (wishlist_id, product_id)
         DO UPDATE SET target_price = EXCLUDED.target_price`, // thêm lại = cập nhật giá mong muốn
        wishlistID, productID, target)
    return err
}

// ownsWishlist trả true nếu wishlist thuộc user (DEC-TRACK-12). Handler gọi trước mọi thao tác theo {id}.
func (r *Repo) ownsWishlist(ctx context.Context, userID, wishlistID int64) (bool, error) {
    var ok bool
    err := r.pool.QueryRow(ctx,
        `SELECT EXISTS(SELECT 1 FROM wishlist WHERE id = $1 AND user_id = $2)`,
        wishlistID, userID).Scan(&ok)
    return ok, err
}
```

### Handler (Go) - kiểm chủ sở hữu

```go
// services/track/internal/api/wishlist.go
func (h *Handler) HandleAddItem(w http.ResponseWriter, req *http.Request) {
    userID := auth.UserID(req.Context())
    wid, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
    if err != nil {
        writeErr(w, http.StatusBadRequest, "invalid wishlist id")
        return
    }
    owns, err := h.repo.OwnsWishlist(req.Context(), userID, wid)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    if !owns {
        writeErr(w, http.StatusNotFound, "wishlist not found") // 404, không 403 (DEC-TRACK-12)
        return
    }
    var body struct {
        ProductID   int64  `json:"product_id"`
        TargetPrice *int64 `json:"target_price"`
    }
    if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid body")
        return
    }
    if err := h.repo.AddItem(req.Context(), wid, body.ProductID, body.TargetPrice); err != nil {
        if isFKViolation(err) { // product_id chưa có trong tracked_product
            writeErr(w, http.StatusBadRequest, "product not tracked")
            return
        }
        writeErr(w, http.StatusInternalServerError, "internal error")
        return
    }
    w.WriteHeader(http.StatusOK) // idempotent: 200 dù mới hay đã có
}
```

---

## §4 - Acceptance criteria

1. `POST /v1/wishlists {name:"Cần mua"}` trả `201` + `{id, name}`; dòng `wishlist` có `user_id` của caller.
2. `GET /v1/wishlists` chỉ trả wishlist của caller; wishlist của user khác KHÔNG xuất hiện.
3. `POST /v1/wishlists/{id}/items {product_id, target_price:89000}` trả `200`; `wishlist_item` lưu `target_price=89000` (int64).
4. Thêm cùng `product_id` lần hai với `target_price` khác -> `200`, một dòng, `target_price` đã cập nhật (idempotent).
5. Thêm `product_id` chưa có trong `tracked_product` -> `400` (FK chặn id rác).
6. Mọi route theo `{id}` với wishlist của user khác -> `404` (không `403`, không lộ tồn tại).
7. `DELETE /v1/wishlists/{id}` xóa wishlist và mọi `wishlist_item` (CASCADE); item không còn.
8. `DELETE /v1/wishlists/{id}/items/{product_id}` gỡ đúng một item; các item khác giữ nguyên.
9. `target_price` trong JSON là int64 hoặc `null`; không float, không string.
10. `UNIQUE (wishlist_id, product_id)` chặn hai dòng cùng SKU trong một wishlist qua insert trần.
11. Request thiếu JWT bị gateway chặn; handler không tự xác thực.
12. Metric `wishlist_ops_total{op, status}` tăng đúng theo từng thao tác.

---

## §5 - Kiểm thử (verification)

```go
// services/track/internal/track/wishlist_repo_test.go
func TestAddItem_Idempotent_UpdatesTarget(t *testing.T) {
    r, uid, wid, pid := setupWishlist(t)
    require.NoError(t, r.AddItem(ctx, wid, pid, ptr(int64(99_000))))
    require.NoError(t, r.AddItem(ctx, wid, pid, ptr(int64(79_000)))) // thêm lại
    items := r.listItems(t, wid)
    require.Len(t, items, 1)                          // không nhân đôi
    require.Equal(t, int64(79_000), *items[0].TargetPrice) // giá đã cập nhật
}

func TestAddItem_UnknownProduct_FK(t *testing.T) {
    r, _, wid, _ := setupWishlist(t)
    err := r.AddItem(ctx, wid, 999999, nil)
    require.Error(t, err) // FK tracked_product -> handler map 400
}

// services/track/internal/api/wishlist_test.go
func TestList_ScopedToUser(t *testing.T) {
    h := setupHandler(t)
    widA := createWishlist(t, h, userA, "A")
    createWishlist(t, h, userB, "B")
    rec := doGETAs(t, h, userA, "/v1/wishlists")
    var lists []WishlistDTO
    decode(t, rec, &lists)
    require.Len(t, lists, 1)
    require.Equal(t, widA, lists[0].ID) // không thấy wishlist của userB
}

func TestCrossUser_404(t *testing.T) {
    h := setupHandler(t)
    widB := createWishlist(t, h, userB, "B")
    rec := doPOSTAs(t, h, userA, "/v1/wishlists/"+itoa(widB)+"/items",
        `{"product_id":1}`)
    require.Equal(t, 404, rec.Code) // không 403 (DEC-TRACK-12)
}

func TestDeleteWishlist_Cascade(t *testing.T) {
    h := setupHandler(t)
    wid := createWishlist(t, h, userA, "A")
    addItem(t, h, userA, wid, trackedPID(t, h))
    doDELETEAs(t, h, userA, "/v1/wishlists/"+itoa(wid))
    require.Equal(t, 0, countItems(t, h, wid)) // CASCADE xóa item
}

func TestTargetPrice_Int64InJSON(t *testing.T) {
    h := setupHandler(t)
    wid := createWishlist(t, h, userA, "A")
    pid := trackedPID(t, h)
    addItemWithTarget(t, h, userA, wid, pid, 89_000)
    rec := doGETAs(t, h, userA, "/v1/wishlists/"+itoa(wid)+"/items")
    require.Contains(t, rec.Body.String(), `"target_price":89000`) // số nguyên, không "89000.0"
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0002_wishlist.sql` (2 bảng + FK + UNIQUE + CASCADE) -> `wishlist_repo.go` (5 hàm, mỗi hàm theo `{id}` đi qua `OwnsWishlist`) -> `wishlist.go` (5 handler) -> đăng ký route trong `router.go` sau JWT middleware (FR-INFRA-001) -> tests. Handler dùng `http.ServeMux` Go 1.22 với `req.PathValue`. Map lỗi FK của `pgx` (`SQLSTATE 23503`) sang `400 product not tracked`. Kiểm chủ sở hữu là bước đầu tiên của mọi handler theo `{id}` - không có ngoại lệ.

---

## §7 - Phụ thuộc

- **FR-TRACK-001** - `tracked_product` phải có SKU đã track trước khi thêm vào wishlist (FK `product_id`).
- **FR-INFRA-002** - bảng `app_user` cho FK `user_id`.
- **FR-INFRA-001 (gateway)** - gắn JWT auth và `user_id` vào context.
- **FR-TRACK-003 (sibling)** - alert_rule type `price_below` đọc `target_price` khi user đặt giá mong muốn.
- **FR-WEB-004 (downstream)** - UI quản lý wishlist tiêu thụ các route này.
- Lib: `pgx`, `encoding/json`, `net/http`.

---

## §8 - Payload ví dụ

### Tạo wishlist + thêm item

```
curl -s -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
  -d '{"name":"Chờ sale 12.12"}' "https://api.sandeal.vn/v1/wishlists"
# -> {"id": 42, "name": "Chờ sale 12.12"}

curl -s -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
  -d '{"product_id": 90112, "target_price": 79000}' \
  "https://api.sandeal.vn/v1/wishlists/42/items"
# -> 200
```

### GET items (200)

```json
[
  { "product_id": 90112, "target_price": 79000, "added_at": "2026-06-28T03:10:00Z" },
  { "product_id": 90118, "target_price": null,  "added_at": "2026-06-28T03:12:00Z" }
]
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Wishlist chia sẻ / công khai (share link cho người khác xem) - thêm khi có nhu cầu social; cần lại mô hình phân quyền.
- Sắp xếp / ghim item trong wishlist (cột `position`) - tối ưu UI giai đoạn sau.
- Giới hạn số wishlist/số item theo tier (free vs Premium) - gắn vào FR-BILL khi có gating.
- Tự tạo alert_rule price_below khi user đặt target_price - để FR-TRACK-003 quyết định liên kết, tránh ghép cứng ở đây.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Truy cập wishlist user khác | OwnsWishlist false | 404 | Kiểm chủ sở hữu mọi route {id} (DEC-TRACK-12) |
| 403 vô tình lộ tồn tại | code review | rò rỉ id | Luôn trả 404 cho id không thuộc caller |
| Thêm product_id rác | FK 23503 | 400 product not tracked | Buộc SKU đã qua FR-TRACK-001 |
| Thêm trùng SKU | ON CONFLICT | 200 no-op / cập nhật target | Idempotent (DEC-TRACK-13) |
| target_price float | CHECK + kiểu int64 | từ chối / sai kiểu | BIGINT VND, JSON int64 |
| Xóa wishlist còn item | ON DELETE CASCADE | item tự xóa | CASCADE ở FK |
| id wishlist không phải số | ParseInt lỗi | 400 | Frontend gửi id số |
| Race thêm item cùng SKU | ON CONFLICT | một dòng | Theo thiết kế (idempotent) |
| user_id giả từ body | lấy từ JWT | bỏ qua | Không nhận user_id từ client |

---

## §11 - Ghi chú

- `wishlist` cho người dùng tổ chức ý định mua sắm; `target_price` là cầu nối sang luồng cảnh báo giá (FR-TRACK-003).
- Kiểm chủ sở hữu trả 404 (không 403) là rào chống IDOR + chống liệt kê id tuần tự - một chi tiết bảo mật dễ bỏ sót với khóa BIGSERIAL.
- Thêm item idempotent (UNIQUE + ON CONFLICT DO UPDATE) làm "bấm lại" chỉ cập nhật giá mong muốn, không nhân đôi.
- FK `product_id -> tracked_product` giữ wishlist luôn gắn SKU thật đang được theo dõi giá, không có dòng trỏ hư không.
- `target_price` là int64 VND suốt đường DB tới JSON, không có bước float nào để tránh sai số khi engine alert so giá.
- Khi mở SEA, `target_price` vẫn int64 theo minor unit từng nước (gắn currency vào tracked_product); schema không đổi.

---

*Hết FR-TRACK-002. Status: ready_to_implement (mục tiêu audit 10/10).*
