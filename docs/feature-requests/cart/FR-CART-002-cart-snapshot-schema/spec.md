---
id: FR-CART-002
title: "Schema cart_snapshot + cart_item - ảnh chụp giỏ hàng nhận từ extension (qty, unit_price BIGINT VND, shop_id), gắn user + platform, là đầu vào optimizer; chỉ dữ liệu tối thiểu hóa, không cookie/token"
module: CART
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-EXT-003, FR-EXT-002, FR-CART-001, FR-CART-003, FR-INFRA-002]
depends_on: [FR-EXT-003]
blocks: [FR-CART-003]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model cart_snapshot + cart_item: qty, unit_price, shop_id)"
  - "docs/... §3.2 (extension gửi dữ liệu tối thiểu hóa: productId/price/qty, KHÔNG cookie/token), §5.4 (tối thiểu hóa dữ liệu)"
source_decisions:
  - "DEC-CART-07: cart_snapshot là một lần chụp giỏ của một user trên một platform tại captured_at; cart_item là các dòng SKU trong ảnh chụp đó (qty, unit_price, shop_id)"
  - "DEC-CART-08: unit_price lưu BIGINT VND (đồng nhất DEC-PRICE-05); qty INTEGER > 0; KHÔNG float"
  - "DEC-CART-09: cart_item CHỈ chứa dữ liệu tối thiểu hóa từ FR-EXT-003 (product ref + qty + unit_price + shop_id); TUYỆT ĐỐI không cột cookie/token/header - cam kết niềm tin lõi"
  - "DEC-CART-10: cart_snapshot.user_id gắn từ JWT (FR-AUTH-002) phía server, KHÔNG nhận từ payload extension (chống giả mạo chủ sở hữu); ghi snapshot là user-scoped"
  - "DEC-CART-11: ghi snapshot idempotent theo client-cung-cấp snapshot_ref (UUID) trong cửa sổ ngắn để retry extension không tạo bản trùng; ảnh chụp là bất biến sau khi ghi"
  - "DEC-CART-12: product trong cart_item tham chiếu tracked_product khi đã biết; nếu chưa track, lưu platform_item_id thô + shop_id để optimizer vẫn tính được, không chặn ghi"

language: "PostgreSQL 16; migration golang-migrate (FR-INFRA-002); service Go 1.22 (cart-svc) cho API nhận snapshot"
service: shopass/services/cart/
new_files:
  - db/migrations/0011_cart_snapshot.up.sql
  - db/migrations/0011_cart_snapshot.down.sql
  - services/cart/internal/cart/snapshot_types.go
  - services/cart/internal/cart/snapshot_repo.go
  - services/cart/internal/api/snapshot.go
  - services/cart/internal/cart/snapshot_repo_test.go
  - services/cart/internal/api/snapshot_test.go
modified_files:
  - services/cart/internal/api/router.go         # đăng ký POST /v1/cart/snapshot
allowed_tools:
  - file_read: services/cart/**
  - file_read: db/migrations/**
  - file_write: services/cart/**
  - file_write: db/migrations/**
  - bash: cd services/cart && go test ./...
disallowed_tools:
  - thêm cột/nhận trường cookie/token/header/session vào cart_snapshot hay cart_item (vi phạm DEC-CART-09 - phá cam kết niềm tin lõi + PDPL)
  - lấy user_id từ payload extension thay vì JWT (vi phạm DEC-CART-10, giả mạo chủ sở hữu)
  - lưu unit_price dạng float (vi phạm DEC-CART-08, sai số tiền tệ)
  - chặn ghi snapshot khi SKU chưa track (vi phạm DEC-CART-12, mất khả năng tối ưu giỏ tạm)

effort_hours: 5
sub_tasks:
  - "0.5h: 0011_cart_snapshot.up/down.sql - bảng cart_snapshot + cart_item + FK + CHECK qty>0/unit_price>0"
  - "0.75h: snapshot_types.go - CartSnapshot + CartItem (BIGINT) + DTO nhận từ extension"
  - "1.0h: snapshot_repo.go - InsertSnapshot (transaction: snapshot + items), GetSnapshot, scope user_id"
  - "1.0h: snapshot.go - handler POST /v1/cart/snapshot: user_id từ JWT, validate payload tối thiểu, idempotent theo snapshot_ref"
  - "1.0h: snapshot_repo_test.go - insert snapshot + items, qty>0/unit_price int64, cross-user không đọc được, idempotent"
  - "0.75h: snapshot_test.go - user_id từ JWT không từ payload; payload chứa cookie/token bị từ chối/bỏ; SKU chưa track vẫn ghi"

risk_if_skipped: "cart_snapshot là cách backend nhận ảnh chụp giỏ hàng mà extension đọc được (FR-EXT-002/003) để optimizer (FR-CART-003) tính tổ hợp voucher tốt nhất - không có nó thì optimizer không có giỏ để tối ưu, và toàn bộ tính năng 'tối ưu giỏ hàng' (persona Linh, §1.2) mất đầu vào. Đây là FR chạm dữ liệu nhạy cảm: giỏ hàng phản ánh ý định mua sắm. Nguy hiểm nhất là nếu schema lỡ nhận/lưu cookie/token (dù chỉ một cột) thì phá vỡ cam kết tối thiểu hóa dữ liệu (§5.4) và token-không-rời-client (§3.2) - đụng thẳng PDPL (chế tài tới 5% doanh thu) và giết định vị niềm tin hậu-Honey. Nếu lấy user_id từ payload extension thay vì JWT thì một payload giả mạo gắn giỏ vào user khác (chiếm/làm loạn dữ liệu người khác). Lưu unit_price float gây sai số khi optimizer cộng giá trị đơn so với min_spend voucher. Chặn ghi khi SKU chưa track làm mất khả năng tối ưu giỏ tạm thời (nhiều SKU mới chưa có trong tracked_product)."
---

## §1 - Mô tả (BCP-14 normative)

Service CART **MUST** định nghĩa schema `cart_snapshot` + `cart_item` và API nhận ảnh chụp giỏ hàng từ extension, chứa CHỈ dữ liệu tối thiểu hóa (qty, unit_price VND, shop_id), gắn chủ sở hữu từ JWT, làm đầu vào cho optimizer. Hợp đồng:

1. **MUST** định nghĩa `cart_snapshot (id BIGSERIAL PK, user_id BIGINT REFERENCES app_user(id), platform_id SMALLINT REFERENCES platform(id), snapshot_ref UUID, captured_at TIMESTAMPTZ NOT NULL DEFAULT now())`.
2. **MUST** định nghĩa `cart_item (id BIGSERIAL PK, cart_snapshot_id BIGINT REFERENCES cart_snapshot(id) ON DELETE CASCADE, product_id BIGINT REFERENCES tracked_product(id), platform_item_id TEXT, shop_id TEXT, qty INTEGER NOT NULL, unit_price BIGINT NOT NULL)`.
3. **MUST** lưu `unit_price` dạng `BIGINT` VND và `qty` `INTEGER` (DEC-CART-08): `qty > 0`, `unit_price > 0` qua CHECK. KHÔNG dùng float.
4. `cart_snapshot` và `cart_item` **MUST NOT** chứa bất kỳ cột nào lưu cookie, session token, header xác thực, hay credential (DEC-CART-09). Chỉ chứa dữ liệu tối thiểu hóa do FR-EXT-003 cho qua (tham chiếu sản phẩm + qty + unit_price + shop_id).
5. API `POST /v1/cart/snapshot` **MUST** lấy `user_id` từ JWT (do gateway gắn, FR-AUTH-002/FR-INFRA-001), KHÔNG từ payload extension (DEC-CART-10); payload có trường tên user_id/owner bị bỏ qua.
6. Handler **MUST** từ chối (hoặc loại bỏ) payload chứa bất kỳ trường nào tên cookie/token/session/authorization/header (defense in depth cùng FR-EXT-003) - không bao giờ ghi credential vào DB.
7. Ghi snapshot **MUST** trong một transaction: tạo `cart_snapshot` rồi chèn tất cả `cart_item`; lỗi giữa chừng rollback toàn bộ (ảnh chụp nguyên tử).
8. Ghi snapshot **MUST** idempotent theo `snapshot_ref` (UUID client cung cấp) trong cửa sổ retry (DEC-CART-11): cùng `snapshot_ref` gửi lại trả snapshot đã ghi, không tạo bản trùng. `UNIQUE (user_id, snapshot_ref)`.
9. `cart_item.product_id` **MUST** tham chiếu `tracked_product` khi SKU đã biết; nếu chưa track, **MUST** vẫn ghi với `product_id` NULL + `platform_item_id` thô + `shop_id` (DEC-CART-12) - không chặn ghi để optimizer vẫn tính được giỏ tạm.
10. Đọc snapshot (`GetSnapshot`) **MUST** scope theo `user_id`: user chỉ đọc snapshot của chính mình; truy cập snapshot user khác trả `404` (không phân biệt không-tồn-tại với của-người-khác).
11. Ảnh chụp **MUST** bất biến sau khi ghi (DEC-CART-11): không cập nhật cart_item của snapshot đã ghi; giỏ đổi thì tạo snapshot mới (lịch sử giỏ).
12. **SHOULD** phát OTel `cart_snapshot_written_total{platform_id}` và `cart_item_count` (histogram số item/giỏ) để theo dõi kích thước giỏ.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao tách snapshot khỏi item (DEC-CART-07)?** Một giỏ hàng là một tập SKU tại một thời điểm. `cart_snapshot` là tiêu đề (ai, sàn nào, lúc nào); `cart_item` là các dòng. Tách ra cho phép một user có nhiều ảnh chụp theo thời gian (lịch sử giỏ) và optimizer tính trên một ảnh chụp cụ thể. Đây là mẫu header-detail kinh điển.

**Vì sao tuyệt đối không cột credential (DEC-CART-09)?** Định vị lõi của SănDeal hậu-Honey là tối thiểu hóa dữ liệu (§5.4) và token-không-rời-client (§3.2). Nếu schema có dù chỉ một cột lưu cookie/token thì phá vỡ cam kết đó - một sự cố rò rỉ là thảm họa PDPL. Schema chỉ chứa dữ liệu giỏ đã tối thiểu hóa. Đây là ranh giới thiết kế, không chỉ quy ước - không có cột thì không thể lỡ lưu.

**Vì sao user_id từ JWT không từ payload (DEC-CART-10)?** Extension chạy trên máy người dùng - payload có thể bị giả mạo. Nếu tin `user_id` trong payload thì một payload độc gắn giỏ vào user khác (làm loạn/chiếm dữ liệu). Lấy `user_id` từ JWT đã verify (gateway) là nguồn chủ sở hữu đáng tin duy nhất. Payload chỉ mang dữ liệu giỏ, không mang danh tính.

**Vì sao idempotent theo snapshot_ref (DEC-CART-11)?** Extension có retry (mạng chập chờn); cùng một lần chụp giỏ có thể gửi hai lần. `snapshot_ref` (UUID client sinh một lần cho một lần chụp) + UNIQUE làm gửi lại trả snapshot cũ thay vì tạo bản trùng. Ảnh chụp bất biến nên trả lại bản đã ghi là đúng.

**Vì sao vẫn ghi khi SKU chưa track (DEC-CART-12)?** Giỏ người dùng thường có SKU SănDeal chưa từng thấy (chưa có trong `tracked_product`). Nếu chặn ghi vì FK thì mất khả năng tối ưu giỏ đó. Cho `product_id` NULL + giữ `platform_item_id` thô để optimizer vẫn tính giá trị đơn (đủ cho min_spend voucher); track sau (FR-TRACK-001) khi cần lịch sử giá.

**Vì sao ảnh chụp bất biến (§1 #11)?** Giỏ là sự kiện tại một thời điểm. Sửa ảnh chụp cũ làm mất ý nghĩa "giỏ lúc đó". Giỏ đổi thì tạo snapshot mới - giữ lịch sử và tránh ghi đè nhầm.

---

## §3 - Hợp đồng API / DDL

### Migration (golang-migrate)

```sql
-- db/migrations/0011_cart_snapshot.up.sql
CREATE TABLE cart_snapshot (
  id           BIGSERIAL   PRIMARY KEY,
  user_id      BIGINT      NOT NULL REFERENCES app_user(id),
  platform_id  SMALLINT    NOT NULL REFERENCES platform(id),
  snapshot_ref UUID        NOT NULL,
  captured_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, snapshot_ref)        -- idempotent retry (DEC-CART-11)
);

CREATE TABLE cart_item (
  id               BIGSERIAL PRIMARY KEY,
  cart_snapshot_id BIGINT    NOT NULL REFERENCES cart_snapshot(id) ON DELETE CASCADE,
  product_id       BIGINT    REFERENCES tracked_product(id),  -- NULL nếu chưa track (DEC-CART-12)
  platform_item_id TEXT,                                      -- giữ ref thô khi product_id NULL
  shop_id          TEXT,
  qty              INTEGER   NOT NULL CHECK (qty > 0),
  unit_price       BIGINT    NOT NULL CHECK (unit_price > 0), -- VND
  -- phải có ít nhất một cách định danh SKU
  CONSTRAINT item_identified CHECK (product_id IS NOT NULL OR platform_item_id IS NOT NULL)
);
CREATE INDEX idx_ci_snapshot ON cart_item (cart_snapshot_id);

-- db/migrations/0011_cart_snapshot.down.sql
DROP TABLE cart_item;
DROP TABLE cart_snapshot;
```

### Types + DTO (snapshot_types.go)

```go
// services/cart/internal/cart/snapshot_types.go
type CartItem struct {
    ProductID      *int64  `db:"product_id"`        // NULL nếu chưa track
    PlatformItemID *string `db:"platform_item_id"`
    ShopID         *string `db:"shop_id"`
    Qty            int32   `db:"qty"`
    UnitPrice      int64   `db:"unit_price"`        // VND
}
type CartSnapshot struct {
    ID          int64     `db:"id"`
    UserID      int64     `db:"user_id"`            // gắn từ JWT, KHÔNG từ payload
    PlatformID  int16     `db:"platform_id"`
    SnapshotRef uuid.UUID `db:"snapshot_ref"`
    CapturedAt  time.Time `db:"captured_at"`
    Items       []CartItem
}

// DTO nhận từ extension - CHÚ Ý: KHÔNG có user_id, KHÔNG có cookie/token (DEC-CART-09/10)
type SnapshotRequest struct {
    PlatformID  int16     `json:"platform_id"`
    SnapshotRef uuid.UUID `json:"snapshot_ref"`
    Items       []struct {
        PlatformItemID *string `json:"platform_item_id"`
        ShopID         *string `json:"shop_id"`
        Qty            int32   `json:"qty"`
        UnitPrice      int64   `json:"unit_price"` // VND int64
    } `json:"items"`
}
```

### Handler (snapshot.go) - user_id từ JWT, không từ payload

```go
// services/cart/internal/api/snapshot.go
func (h *Handler) HandleSnapshot(w http.ResponseWriter, req *http.Request) {
    userID := auth.UserIDFromContext(req.Context()) // từ JWT (gateway), KHÔNG từ payload (DEC-CART-10)
    if userID == 0 {
        writeErr(w, http.StatusUnauthorized, "unauthenticated"); return
    }
    var in SnapshotRequest
    if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
        writeErr(w, http.StatusBadRequest, "invalid payload"); return
    }
    if err := validateSnapshot(in); err != nil {     // qty>0, unit_price>0, có định danh SKU
        writeErr(w, http.StatusBadRequest, err.Error()); return
    }
    snap, err := h.repo.InsertSnapshot(req.Context(), userID, in) // idempotent theo (user_id, snapshot_ref)
    if err != nil {
        writeErr(w, http.StatusInternalServerError, "internal error"); return
    }
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(map[string]any{"snapshot_id": snap.ID, "item_count": len(snap.Items)})
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch (up/down); `cart_snapshot` + `cart_item` tồn tại với CHECK `qty>0`, `unit_price>0`, `item_identified`, và `UNIQUE(user_id, snapshot_ref)`.
2. Grep `cart_snapshot`/`cart_item` schema + types + DTO: KHÔNG có cột/trường nào tên cookie/token/session/authorization/header.
3. INSERT `qty<=0` hoặc `unit_price<=0` -> lỗi CHECK; `unit_price` round-trip int64 (không float).
4. `POST /v1/cart/snapshot` lấy `user_id` từ JWT context; payload có trường `user_id`/`owner` bị bỏ qua (không ghi đè chủ sở hữu).
5. Payload chứa trường tên cookie/token/... bị từ chối hoặc loại bỏ; DB không bao giờ có giá trị đó.
6. Ghi snapshot là transaction: snapshot + items cùng commit; lỗi giữa chừng rollback (không snapshot mồ côi không item).
7. Cùng `snapshot_ref` (cùng user) gửi hai lần -> trả snapshot đã ghi, một bản (idempotent), không nhân item.
8. SKU chưa track (không product_id) vẫn ghi với `platform_item_id` + `shop_id`; CHECK `item_identified` chấp nhận.
9. `GetSnapshot` của user A không trả snapshot của user B (scope user_id; `404` cho cross-user).
10. Snapshot đã ghi bất biến: không có đường cập nhật cart_item của snapshot cũ (chỉ tạo mới).
11. `go test ./...` xanh.

---

## §5 - Kiểm thử (verification)

```go
// services/cart/internal/cart/snapshot_repo_test.go
func TestInsertSnapshot_TransactionalWithItems(t *testing.T) {
    r, uid := setupCart(t)
    snap, err := r.InsertSnapshot(ctx, uid, reqWithItems(2)) // 2 item
    require.NoError(t, err)
    require.Len(t, snap.Items, 2)
    require.Equal(t, uid, snap.UserID) // chủ sở hữu từ JWT
}

func TestInsertSnapshot_Idempotent(t *testing.T) {
    r, uid := setupCart(t)
    ref := uuid.New()
    in := reqWithRef(ref, 2)
    s1, _ := r.InsertSnapshot(ctx, uid, in)
    s2, _ := r.InsertSnapshot(ctx, uid, in) // cùng (user, snapshot_ref)
    require.Equal(t, s1.ID, s2.ID)          // một bản
    require.Equal(t, 2, countItems(t, r, s1.ID)) // không nhân item
}

func TestInsertSnapshot_UntrackedSku_StillWrites(t *testing.T) {
    r, uid := setupCart(t)
    in := reqUntracked("ITEM-999", "shopX", 1, 50_000) // không product_id
    snap, err := r.InsertSnapshot(ctx, uid, in)
    require.NoError(t, err) // CHECK item_identified chấp nhận (DEC-CART-12)
    require.Nil(t, snap.Items[0].ProductID)
    require.Equal(t, "ITEM-999", *snap.Items[0].PlatformItemID)
}

func TestGetSnapshot_CrossUser_404(t *testing.T) {
    r, uidA := setupCart(t)
    snap, _ := r.InsertSnapshot(ctx, uidA, reqWithItems(1))
    uidB := makeUser(t, r)
    _, err := r.GetSnapshot(ctx, uidB, snap.ID) // user B đọc snapshot của A
    require.ErrorIs(t, err, ErrNotFound)
}

func TestCheck_QtyAndPrice(t *testing.T) {
    r, uid := setupCart(t)
    _, err := r.InsertSnapshot(ctx, uid, reqBadQty(0))
    require.Error(t, err) // CHECK qty > 0
}
```

```go
// services/cart/internal/api/snapshot_test.go
func TestHandler_UserIDFromJWT_NotPayload(t *testing.T) {
    h, uid := setupHandler(t)
    body := `{"platform_id":1,"snapshot_ref":"...","user_id":99999,"items":[{"platform_item_id":"A","qty":1,"unit_price":50000}]}`
    rec := doPostWithJWT(t, h, "/v1/cart/snapshot", body, uid)
    require.Equal(t, 201, rec.Code)
    snap := lastSnapshot(t, h, uid)
    require.Equal(t, uid, snap.UserID) // user_id = JWT, KHÔNG phải 99999 trong payload
}

func TestHandler_RejectsCredentialFields(t *testing.T) {
    h, uid := setupHandler(t)
    body := `{"platform_id":1,"snapshot_ref":"...","cookie":"SPC=abc","token":"xyz","items":[{"platform_item_id":"A","qty":1,"unit_price":50000}]}`
    doPostWithJWT(t, h, "/v1/cart/snapshot", body, uid)
    // khẳng định không giá trị credential nào lọt vào DB
    require.False(t, anyColumnContains(t, h, "abc"))
    require.False(t, anyColumnContains(t, h, "xyz"))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `0011_cart_snapshot.up/down.sql` (hai bảng + CHECK + UNIQUE + index) -> `snapshot_types.go` (CartSnapshot/CartItem BIGINT + DTO không-credential) -> `snapshot_repo.go` (`InsertSnapshot` transaction, `GetSnapshot` scope user) -> `snapshot.go` (handler: user_id từ JWT, validate, idempotent) -> đăng ký route -> tests. Migration 0011 nối tiếp 0010 (FR-CART-001). Handler nằm sau JWT middleware của gateway (FR-INFRA-001); `auth.UserIDFromContext` đọc sub đã verify. DTO `SnapshotRequest` cố tình KHÔNG có trường user_id/cookie/token - decode JSON bỏ qua khóa thừa, và test khẳng định không giá trị credential nào vào DB (defense in depth cùng FR-EXT-003).

---

## §7 - Phụ thuộc

- **FR-EXT-003** - pipeline tối thiểu hóa dữ liệu client quyết định trường nào được gửi (productId/price/qty, không cookie/token); FR-CART-002 nhận đúng tập đó (depends_on cứng).
- **FR-EXT-002 (nguồn)** - content script đọc giỏ; dữ liệu đi qua FR-EXT-003 rồi tới API này.
- **FR-INFRA-002** - bảng `platform` + `app_user` + `tracked_product` (FK) phải có trước; migration framework.
- **FR-CART-001 (sibling)** - voucher_catalog; cùng nuôi optimizer FR-CART-003.
- **FR-CART-003 (downstream)** - optimizer đọc một `cart_snapshot` làm giỏ để tối ưu voucher.
- Extension/lib: `pgx`, `google/uuid`; golang-migrate.

---

## §8 - Payload ví dụ

### Extension gửi ảnh chụp giỏ (tối thiểu, KHÔNG credential, KHÔNG user_id)

```json
{
  "platform_id": 1,
  "snapshot_ref": "2f1c9b7a-1e2d-4c3b-9a0f-7d6e5c4b3a21",
  "items": [
    { "platform_item_id": "90112", "shop_id": "shopA", "qty": 1, "unit_price": 89000 },
    { "platform_item_id": "77310", "shop_id": "shopB", "qty": 2, "unit_price": 245000 }
  ]
}
```

### Phản hồi (201)

```json
{ "snapshot_id": 5012, "item_count": 2 }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Gắn `product_id` tự động bằng cách track ngầm SKU chưa biết (gọi FR-TRACK-001) lúc nhận snapshot - cân nhắc cân bằng với chi phí scraping.
- Dọn snapshot cũ (retention giỏ) - thêm policy khi dữ liệu giỏ phình; ảnh chụp là dữ liệu cá nhân nên gắn DSAR (FR-COMPLY-003).
- Gộp item trùng SKU trong một giỏ (cùng product, cộng qty) - quyết khi gặp dữ liệu thật.
- Đa tiền tệ khi mở SEA - giữ unit_price BIGINT theo minor unit từng nước.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Cột/trường credential trong giỏ | grep schema + test | rò rỉ token -> thảm họa PDPL | KHÔNG cột credential (DEC-CART-09) |
| user_id từ payload | handler test | giả mạo chủ sở hữu | Lấy từ JWT (DEC-CART-10) |
| unit_price float | review + test | sai số so min_spend | BIGINT VND (DEC-CART-08) |
| Chặn ghi khi SKU chưa track | untracked test | mất tối ưu giỏ tạm | product_id NULL + platform_item_id (DEC-CART-12) |
| Snapshot mồ côi không item | transaction test | giỏ rỗng giả | Ghi trong transaction (§1 #7) |
| Retry tạo bản trùng | idempotent test | giỏ nhân đôi | UNIQUE(user_id, snapshot_ref) (DEC-CART-11) |
| Đọc giỏ user khác | cross-user test | rò rỉ ý định mua | Scope user_id, 404 (§1 #10) |
| Sửa ảnh chụp cũ | review | mất ý nghĩa lịch sử | Ảnh chụp bất biến (§1 #11) |
| qty/unit_price <= 0 | DB CHECK | giỏ vô nghĩa | CHECK qty>0, unit_price>0 |

---

## §11 - Ghi chú

- `cart_snapshot` là cách backend nhận giỏ extension đọc được để optimizer (FR-CART-003) tối ưu voucher - đầu vào của tính năng tối ưu giỏ (persona Linh).
- Tuyệt đối không cột credential là ranh giới thiết kế, không chỉ quy ước - không có cột thì không thể lỡ lưu cookie/token, giữ cam kết tối thiểu hóa (§5.4) và token-không-rời-client (§3.2).
- user_id từ JWT (không từ payload) chống giả mạo chủ sở hữu - payload chỉ mang dữ liệu giỏ, không mang danh tính.
- Idempotent theo snapshot_ref cho retry extension không tạo giỏ trùng; ảnh chụp bất biến.
- Cho ghi SKU chưa track (product_id NULL) giữ khả năng tối ưu giỏ tạm - track sau khi cần lịch sử.
- Header-detail (snapshot + item) cho phép lịch sử giỏ và tối ưu trên một ảnh chụp cụ thể.

---

*Hết FR-CART-002. Status: ready_to_implement (mục tiêu audit 10/10).*
