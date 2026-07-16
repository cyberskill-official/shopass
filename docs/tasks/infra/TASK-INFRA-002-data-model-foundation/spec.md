---
id: TASK-INFRA-002
title: "Data-model foundation - migration framework (golang-migrate) + bảng `platform` + cột lõi `app_user` + quy ước đặt tên/migration"
module: INFRA
priority: MUST
status: done
verify: T
phase: P0
milestone: P0 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-INFRA-001, TASK-INFRA-005, TASK-AUTH-001, TASK-AUTH-003, TASK-PRICE-001]
depends_on: []
blocks: [TASK-AFFIL-001, TASK-AUTH-001, TASK-CART-001, TASK-COMPLY-001, TASK-INFRA-005, TASK-NOTIF-001, TASK-PRICE-001]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model: platform, app_user, DDL nền tảng)"
  - "docs/... §3.8 (NFR khả năng mở rộng), §2 (đa sàn, per-country)"
source_decisions:
  - "DEC-INFRA-06: migration framework dùng golang-migrate; mỗi migration cặp up/down, đánh số tăng dần, bất biến sau khi merge"
  - "DEC-INFRA-07: bảng `platform` là gốc tham chiếu cho mọi liên kết sàn; code - {'shopee','tiktok','lazada'}, country ISO-3166 alpha-2"
  - "DEC-INFRA-08: `app_user` định nghĩa cột lõi ở migration nền (id BIGSERIAL, email CITEXT unique, locale 'vi-VN'); TASK-AUTH-001 mở rộng cột bảo mật"
  - "DEC-INFRA-09: quy ước đặt tên - bảng snake_case số ít, khoá chính `id`, timestamp `*_at` kiểu TIMESTAMPTZ DEFAULT now()"
  - "DEC-INFRA-10: extension CITEXT + (sau) timescaledb được bật ở migration 0001 trước mọi bảng phụ thuộc"

language: "PostgreSQL 16; golang-migrate; service Go 1.22 (shared db package)"
service: shopass/db/
new_files:
  - db/migrations/0001_extensions.up.sql
  - db/migrations/0001_extensions.down.sql
  - db/migrations/0002_platform.up.sql
  - db/migrations/0002_platform.down.sql
  - db/migrations/0003_app_user_core.up.sql
  - db/migrations/0003_app_user_core.down.sql
  - db/seed/0001_platform_seed.sql
  - db/internal/migrate/migrate.go
  - db/internal/migrate/migrate_test.go
  - docs/conventions/NAMING-AND-MIGRATIONS.md
modified_files: []
allowed_tools:
  - file_read: db/**
  - file_write: db/**
  - bash: cd db && go test ./...
disallowed_tools:
  - sửa nội dung migration đã merge (vi phạm DEC-INFRA-06; phải thêm migration mới)
  - tạo bảng phụ thuộc CITEXT trước migration 0001_extensions (vi phạm DEC-INFRA-10)
  - dùng SERIAL/INT cho khoá chính bảng có thể lớn (phải BIGSERIAL/BIGINT)

effort_hours: 6
sub_tasks:
  - "0.5h: 0001_extensions - CREATE EXTENSION IF NOT EXISTS citext (+ timescaledb stub cho PRICE)"
  - "0.5h: 0002_platform - bảng platform + CHECK code/country + seed 3 sàn VN"
  - "1.0h: 0003_app_user_core - cột lõi app_user (id, email CITEXT unique, phone, display_name, locale, status, created_at)"
  - "1.0h: migrate.go - runner golang-migrate (up/down/version/force) + healthcheck schema_migrations"
  - "1.0h: migrate_test.go - up sạch từ zero, down rollback, idempotent up lần 2, platform seed đủ 3 dòng"
  - "1.0h: NAMING-AND-MIGRATIONS.md - quy ước đặt tên + quy tắc migration bất biến"
  - "1.0h: kiểm CHECK constraint platform.code/country + unique email CITEXT case-insensitive (test)"

risk_if_skipped: "Không có migration framework thì schema trôi dạt giữa môi trường (dev/staging/prod), không tái lập được, không rollback được - nguồn gốc lỗi production khó truy. Không có bảng `platform` thì mọi liên kết sàn (platform_account, tracked_product, voucher, affiliate) mất gốc FK. Không có cột lõi `app_user` thì AUTH/BILL/TRACK không có chủ thể để gắn. Đây là nền dữ liệu mà 9+ task khác phụ thuộc trực tiếp."
---

## §1 - Mô tả (BCP-14 normative)

Service INFRA **MUST** thiết lập nền dữ liệu PostgreSQL: framework migration golang-migrate có up/down, bảng `platform` làm gốc tham chiếu sàn, cột lõi `app_user`, và một tài liệu quy ước đặt tên/migration. Hợp đồng:

1. Hệ thống **MUST** dùng golang-migrate với mỗi migration là một cặp file `NNNN_name.up.sql` + `NNNN_name.down.sql`, đánh số tăng dần liên tục, bất biến sau khi merge (DEC-INFRA-06). Thay đổi schema về sau là thêm migration mới, KHÔNG sửa file cũ.
2. Migration `0001_extensions` **MUST** chạy đầu tiên: `CREATE EXTENSION IF NOT EXISTS citext` (và stub bật `timescaledb` cho PRICE) trước mọi bảng phụ thuộc kiểu CITEXT (DEC-INFRA-10).
3. Migration `0002_platform` **MUST** tạo bảng `platform (id SMALLINT PK, code TEXT UNIQUE NOT NULL, country TEXT NOT NULL, base_url TEXT, created_at TIMESTAMPTZ DEFAULT now())`.
4. Bảng `platform` **MUST** ràng buộc `code` - {`'shopee'`,`'tiktok'`,`'lazada'`} qua CHECK, và `country` là mã ISO-3166 alpha-2 (hai chữ in hoa) qua CHECK (DEC-INFRA-07).
5. Seed `0001_platform_seed` **MUST** chèn tối thiểu 3 dòng cho thị trường VN: `(1,'shopee','VN',...)`, `(2,'tiktok','VN',...)`, `(3,'lazada','VN',...)`, idempotent qua `ON CONFLICT (code) DO NOTHING`.
6. Migration `0003_app_user_core` **MUST** tạo `app_user` với cột lõi: `id BIGSERIAL PK`, `email CITEXT UNIQUE`, `phone TEXT`, `display_name TEXT`, `locale TEXT NOT NULL DEFAULT 'vi-VN'`, `status TEXT NOT NULL DEFAULT 'active'`, `created_at TIMESTAMPTZ DEFAULT now()` (DEC-INFRA-08).
7. `app_user.email` **MUST** là CITEXT để so sánh không phân biệt hoa thường: `Foo@x.com` và `foo@x.com` xung đột UNIQUE (chống tạo trùng tài khoản qua biến thể hoa thường).
8. TASK-AUTH-001 **MUST** mở rộng `app_user` ở migration sau (thêm `pwd_hash`, `referral_code_id`, ...); task này KHÔNG định nghĩa cột bảo mật để giữ ranh giới module.
9. Runner migration **MUST** hỗ trợ: `Up()` (áp mọi migration chưa chạy), `Down(n)` (rollback n bước), `Version()` (đọc version hiện tại), và an toàn idempotent (chạy `Up` lần hai không lỗi, không đổi schema).
10. Mọi bảng **MUST** theo quy ước đặt tên (DEC-INFRA-09): tên bảng snake_case số ít, khoá chính tên `id`, cột thời gian hậu tố `_at` kiểu TIMESTAMPTZ. Tài liệu `NAMING-AND-MIGRATIONS.md` ghi quy ước này như nguồn sự thật.
11. Khoá chính của bảng có khả năng lớn (`app_user` và các bảng nghiệp vụ) **MUST** là `BIGSERIAL`/`BIGINT`; chỉ bảng nhỏ cố định (`platform`) dùng `SMALLINT`.
12. Runner **SHOULD** ghi log có cấu trúc mỗi bước migration (version từ -> đến, hướng up/down, thời gian) để truy vết khi deploy.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao golang-migrate với migration bất biến (DEC-INFRA-06)? Schema phải tái lập y hệt giữa dev, staging và prod. Migration đánh số, cặp up/down, không sửa sau khi merge cho ta một lịch sử tuyến tính kiểm toán được: mọi môi trường chạy cùng chuỗi và tới cùng trạng thái. Sửa file cũ phá tính tái lập (môi trường đã chạy bản cũ sẽ lệch).

Vì sao `platform` là gốc (DEC-INFRA-07)? SănDeal đa sàn ngay từ thiết kế. `platform_account`, `tracked_product`, `voucher_catalog`, `affiliate_click` đều tham chiếu một sàn. Tập trung định nghĩa sàn vào một bảng (với CHECK code) tránh chuỗi `'shopee'` rải rác và sai chính tả; thêm sàn mới là thêm một dòng, không phải sửa nhiều nơi.

Vì sao email CITEXT (§1 #7)? Người dùng gõ email với hoa thường tùy hứng. Nếu `email` là TEXT thường, `Foo@x.com` và `foo@x.com` là hai bản ghi khác nhau -> một người tạo được nhiều tài khoản, đăng nhập rối. CITEXT làm UNIQUE so sánh case-insensitive, đúng ngữ nghĩa email.

Vì sao tách cột lõi `app_user` khỏi cột bảo mật (DEC-INFRA-08, §1 #8)? INFRA dựng khung dữ liệu; AUTH lo bảo mật. Để cột lõi (id, email, locale, status) ở migration nền cho mọi module gắn vào; để `pwd_hash`/`referral_code_id` cho TASK-AUTH-001. Ranh giới này giữ migration nền không bị kẹt chờ quyết định bảo mật.

Vì sao quy ước đặt tên thành tài liệu (DEC-INFRA-09)? 16 module sẽ tạo hàng chục bảng. Không có quy ước chung thì mỗi module đặt kiểu khác (`userId` vs `user_id`, `createdAt` vs `created_at`), gây ma sát join và migration. Một tài liệu quy ước là hợp đồng để mọi task sau tuân theo.

Vì sao `locale` mặc định `'vi-VN'` (§1 #6)? Thị trường gốc là Việt Nam; phần lớn người dùng là vi-VN. Đặt mặc định đúng giảm boilerplate và là điểm neo cho i18n và per-country gating (TASK-INFRA-005) khi mở SEA.

---

## §3 - Hợp đồng API / DDL

### Migrations

```sql
-- db/migrations/0001_extensions.up.sql
CREATE EXTENSION IF NOT EXISTS citext;
-- timescaledb được bật khi triển khai PRICE (TASK-PRICE-002); stub ở đây để thứ tự rõ ràng:
-- CREATE EXTENSION IF NOT EXISTS timescaledb;

-- db/migrations/0002_platform.up.sql
CREATE TABLE platform (
  id         SMALLINT     PRIMARY KEY,
  code       TEXT         UNIQUE NOT NULL
               CHECK (code IN ('shopee','tiktok','lazada')),
  country    TEXT         NOT NULL
               CHECK (country ~ '^[A-Z]{2}$'),     -- ISO-3166 alpha-2
  base_url   TEXT,
  created_at TIMESTAMPTZ  DEFAULT now()
);

-- db/migrations/0003_app_user_core.up.sql
CREATE TABLE app_user (
  id           BIGSERIAL    PRIMARY KEY,
  email        CITEXT       UNIQUE,                 -- case-insensitive unique
  phone        TEXT,
  display_name TEXT,
  locale       TEXT         NOT NULL DEFAULT 'vi-VN',
  status       TEXT         NOT NULL DEFAULT 'active',
  created_at   TIMESTAMPTZ  DEFAULT now()
);
-- Cột bảo mật (pwd_hash, referral_code_id) do TASK-AUTH-001 thêm ở migration sau.
```

### Seed (idempotent)

```sql
-- db/seed/0001_platform_seed.sql
INSERT INTO platform (id, code, country, base_url) VALUES
  (1, 'shopee', 'VN', 'https://shopee.vn'),
  (2, 'tiktok', 'VN', 'https://www.tiktok.com'),
  (3, 'lazada', 'VN', 'https://www.lazada.vn')
ON CONFLICT (code) DO NOTHING;
```

### Migration runner (Go)

```go
// db/internal/migrate/migrate.go
type Migrator struct{ m *migrate.Migrate }

func (mg *Migrator) Up() error {
    if err := mg.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
        return err // ErrNoChange (đã mới nhất) là idempotent, không phải lỗi
    }
    return nil
}
func (mg *Migrator) Down(n int) error { return mg.m.Steps(-n) }
func (mg *Migrator) Version() (uint, bool, error) { return mg.m.Version() }
```

---

## §4 - Acceptance criteria

1. `Up()` trên DB rỗng -> mọi migration áp; bảng `platform` và `app_user` tồn tại.
2. `Up()` lần hai -> không lỗi (idempotent), version không đổi (`ErrNoChange` được nuốt).
3. `Down(1)` sau `Up()` -> rollback đúng một bước (bảng của migration cuối biến mất).
4. Seed -> `SELECT count(*) FROM platform` >= 3; có đủ `shopee/tiktok/lazada`.
5. Chạy seed lần hai -> vẫn 3 dòng (ON CONFLICT DO NOTHING).
6. INSERT `platform(code='amazon')` -> lỗi CHECK (code ngoài allowlist).
7. INSERT `platform(country='Vietnam')` -> lỗi CHECK (country không phải alpha-2).
8. INSERT hai `app_user` cùng email khác hoa thường (`A@x.com`, `a@x.com`) -> vi phạm UNIQUE (CITEXT).
9. INSERT `app_user` không set `locale` -> mặc định `'vi-VN'`; không set `status` -> `'active'`.
10. `Version()` trả version hiện tại đúng sau chuỗi up/down.
11. Thứ tự migration: chạy `0002_platform` trước `0001_extensions` thất bại nếu CITEXT chưa bật - xác nhận `0001` đứng trước.
12. `NAMING-AND-MIGRATIONS.md` tồn tại và mô tả quy ước (snake_case số ít, `id`, `*_at`, BIGSERIAL).

---

## §5 - Kiểm thử (verification)

```go
// db/internal/migrate/migrate_test.go
func TestUp_FromZero(t *testing.T) {
    mg := newMigratorFreshDB(t)
    require.NoError(t, mg.Up())
    require.True(t, tableExists(t, "platform"))
    require.True(t, tableExists(t, "app_user"))
}

func TestUp_Idempotent(t *testing.T) {
    mg := newMigratorFreshDB(t)
    require.NoError(t, mg.Up())
    require.NoError(t, mg.Up()) // lần hai không lỗi
}

func TestDown_OneStep(t *testing.T) {
    mg := newMigratorFreshDB(t)
    require.NoError(t, mg.Up())
    require.NoError(t, mg.Down(1))
    require.False(t, tableExists(t, "app_user")) // migration cuối bị rollback
}

func TestPlatformSeed_Idempotent(t *testing.T) {
    db := upWithSeed(t)
    seed(t, db) // chạy seed lần hai
    require.Equal(t, 3, countRows(t, db, "platform"))
}

func TestPlatform_CodeCheck(t *testing.T) {
    db := upWithSeed(t)
    _, err := db.Exec(`INSERT INTO platform(id,code,country) VALUES (9,'amazon','VN')`)
    require.Error(t, err) // CHECK code IN (...)
}

func TestPlatform_CountryCheck(t *testing.T) {
    db := upWithSeed(t)
    _, err := db.Exec(`INSERT INTO platform(id,code,country) VALUES (9,'shopee','Vietnam')`)
    require.Error(t, err) // CHECK country ~ '^[A-Z]{2}$'
}

func TestAppUser_EmailCaseInsensitiveUnique(t *testing.T) {
    db := upWithSeed(t)
    _, err1 := db.Exec(`INSERT INTO app_user(email) VALUES ('A@x.com')`)
    require.NoError(t, err1)
    _, err2 := db.Exec(`INSERT INTO app_user(email) VALUES ('a@x.com')`)
    require.Error(t, err2) // CITEXT unique
}

func TestAppUser_Defaults(t *testing.T) {
    db := upWithSeed(t)
    var locale, status string
    db.QueryRow(`INSERT INTO app_user(email) VALUES ('d@x.com')
                 RETURNING locale, status`).Scan(&locale, &status)
    require.Equal(t, "vi-VN", locale)
    require.Equal(t, "active", status)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự cố định: `0001_extensions` (CITEXT trước) -> `0002_platform` -> `0003_app_user_core` -> seed. Runner bọc golang-migrate; nuốt `ErrNoChange` để `Up` idempotent. Seed chạy sau migration, idempotent qua `ON CONFLICT`. Tài liệu quy ước đặt cạnh migration để mọi task sau tham chiếu. Test chạy trên DB tạm (testcontainers Postgres hoặc instance ephemeral) để kiểm cả CHECK và CITEXT thật, không mock.

---

## §7 - Phụ thuộc

- Không có task phụ thuộc trước (đây là một trong các nền móng P0).
- TASK-AUTH-001 (downstream) - thêm `pwd_hash`, `referral_code_id` vào `app_user` qua migration mới.
- TASK-AUTH-003 / TASK-PRICE-001 / TASK-CART-001 / TASK-AFFIL-001 / TASK-BILL-001 (downstream) - FK tới `platform` và/hoặc `app_user`.
- TASK-INFRA-005 (downstream) - đọc `platform.country` cho per-country gating.
- Hạ tầng: PostgreSQL 16, extension `citext` (và sau là `timescaledb` cho PRICE).

---

## §8 - Payload ví dụ

### Áp migration lúc deploy

```bash
# CI/CD: nâng schema lên bản mới nhất trước khi rollout service
migrate -path db/migrations -database "$DATABASE_URL" up
```

### Trạng thái sau seed

```sql
SELECT id, code, country, base_url FROM platform ORDER BY id;
--  1 | shopee | VN | https://shopee.vn
--  2 | tiktok | VN | https://www.tiktok.com
--  3 | lazada | VN | https://www.lazada.vn
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Phân vùng (partition) `app_user` theo region khi mở SEA - chưa cần ở quy mô P0/P1.
- Soft-delete chuẩn hóa (cột `deleted_at`) cho mọi bảng - gắn vào TASK-AUTH-005 (xóa tài khoản DSAR) và nhân rộng sau.
- Quản lý seed theo môi trường (dev seed khác prod seed) - thêm khi cần dữ liệu mẫu phong phú.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Sửa migration đã merge | diff review/CI | môi trường lệch schema | Cấm sửa; thêm migration mới (DEC-INFRA-06) |
| Tạo bảng CITEXT trước 0001 | migration fail | up đứt giữa chừng | Giữ 0001_extensions đầu chuỗi (§1 #2) |
| Migration nửa chừng (lỗi giữa up) | schema_migrations dirty | deploy kẹt | `force <version>` về điểm sạch rồi chạy lại |
| Email trùng qua biến thể hoa thường | UNIQUE CITEXT | từ chối tạo trùng | Theo thiết kế (§1 #7) |
| Code sàn sai chính tả | CHECK constraint | từ chối insert | Sửa nguồn; chỉ dùng code trong allowlist |
| Country không phải alpha-2 | CHECK constraint | từ chối insert | Chuẩn hóa về ISO-3166 alpha-2 |
| Khoá chính dùng SERIAL bị tràn | giám sát maxval | hết id khi lớn | BIGSERIAL cho bảng lớn (§1 #11) |
| Seed chạy nhiều lần nhân đôi | ON CONFLICT | giữ 3 dòng | Idempotent seed (§1 #5) |
| Down rollback mất dữ liệu prod | review | mất bảng/dữ liệu | Down chỉ ở dev; prod forward-only + backup |

---

## §11 - Ghi chú

- Đây là nền dữ liệu: 9+ task khác FK tới `platform` hoặc `app_user`, hoặc đọc `platform.country`.
- Migration bất biến + đánh số cho lịch sử schema tái lập được giữa mọi môi trường - gốc của deploy tin cậy.
- `platform` tập trung định nghĩa sàn, tránh literal `'shopee'` rải rác và sai chính tả; thêm sàn là thêm một dòng.
- CITEXT trên email chặn tài khoản trùng qua biến thể hoa thường - một lỗi tinh vi nếu dùng TEXT thường.
- Tách cột lõi (INFRA) khỏi cột bảo mật (AUTH) giữ ranh giới module sạch; AUTH mở rộng `app_user` bằng migration riêng.
- Tài liệu quy ước đặt tên là hợp đồng cho mọi task sau, giảm ma sát join/migration khi 16 module cùng tạo bảng.

---

*Hết TASK-INFRA-002. Status: ready_to_implement (mục tiêu audit 10/10).*
