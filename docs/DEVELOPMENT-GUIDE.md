# SănDeal (shopass) - Hướng dẫn phát triển local

Tài liệu này giúp kỹ sư dựng và chạy SănDeal trên máy cá nhân bằng Docker, từng bước. Toàn bộ hạ tầng (Postgres/TimescaleDB) và các service chạy trong Docker Compose, nên bạn không cần cài Postgres hay TimescaleDB thủ công.

Nếu chỉ đọc một mục: chạy "Bắt đầu nhanh" bên dưới là có một stack chạy được trong vài phút.

## 1. Yêu cầu

Bắt buộc:

- Docker (Docker Desktop trên macOS/Windows, hoặc Docker Engine trên Linux) kèm plugin Compose v2. Kiểm tra: `docker compose version`.
- Git.
- make (macOS/Linux có sẵn; Windows dùng WSL2 hoặc chạy trực tiếp các lệnh `docker compose`).

Chỉ cần khi bạn muốn chạy service ngoài Docker hoặc chạy test ở local:

- Go 1.25 (khớp `go.mod`).
- Node.js 22 và npm (extension + web).
- Python 3.11 (service ml). Chạy Prophet đầy đủ cần CmdStan; image ml đã cài sẵn nên trong Docker không phải lo.

## 2. Bắt đầu nhanh (tất cả trong Docker)

```
git clone <repo-url> shopass
cd shopass
make env          # tạo deploy/.env từ mẫu (mặc định dùng được ngay cho local)
make up           # build + chạy: db, migrate, pricesvc, dealsvc, notifsvc, web
make ps           # xem trạng thái các service
```

Mở web tại http://localhost:3000 và API giá tại http://localhost:8081.

Xem toàn bộ vòng lặp lõi chạy end to end (scrape giá -> lưu -> dự báo đáy -> bắn cảnh báo):

```
make smoke
```

Lệnh này seed một sản phẩm demo, chạy scraper, chạy nightly-score của deal, rồi in ra hàng trong `price_snapshot` và `bottom_alert_log`. Lưu ý: scraper mặc định trỏ tới `SHOPEE_BASE_URL=https://shopee.vn`, môi trường local thường bị chặn; để chạy khô, trỏ `SHOPEE_BASE_URL` sang một endpoint fixture trả JSON dạng Shopee, hoặc dùng `scripts/smoke_loop.sh` (script này tự dựng một Shopee giả, cần có Go ở local).

Dừng stack (giữ dữ liệu): `make down`. Xóa luôn dữ liệu: `make clean`.

## 3. Hai chế độ phát triển

Chế độ A - tất cả trong Docker (khuyến nghị để bắt đầu). Sửa code rồi `make up` để build lại image và chạy. Đơn giản, giống môi trường triển khai nhất.

Chế độ B - hạ tầng trong Docker, service chạy local (lặp nhanh khi sửa một service). Dựng riêng database rồi chạy service bằng toolchain local:

```
make dev-db       # chỉ chạy db + migrate

# ví dụ chạy price service ở local, trỏ vào db trong Docker:
cd services/price
DATABASE_URL="postgres://postgres:postgres@localhost:5432/shopass?sslmode=disable" PRICE_ADDR=":8081" go run ./cmd/pricesvc

# chạy web ở local:
cd web && npm install && npm run dev
```

Mỗi service Go có entrypoint trong `services/<svc>/cmd/`. Biến môi trường chính là `DATABASE_URL`; các service khác đọc thêm biến riêng (ví dụ `NOTIFSVC_URL` cho deal, `PRICE_BASE_URL` + `SHOPEE_BASE_URL` cho scrape).

## 4. Các job vòng lặp

Scraper và job dự báo là job chạy theo yêu cầu (không phải service luôn bật):

```
make seed         # seed 1 sản phẩm demo (id 100) + user + luật cảnh báo + 1 forecast "chín"
make scrape       # chạy scraper 1 lần (đặt SCRAPE_SEED=100:555:777)
make forecast     # chạy job dự báo ghi vào price_forecast
make deal-once    # chạy nightly bottom-score 1 lần (thay vì chờ cron 02:00)
make psql         # mở psql để xem dữ liệu
```

Scraper dùng hàng đợi bền vững trên Postgres (bảng `scrape_job`): mỗi lần chạy sẽ đăng ký sản phẩm mới vào hàng đợi rồi rút hết job đến hạn. Sản phẩm "hot" được lên lịch quét lại sau vài phút, "cold" sau khoảng một ngày.

## 5. Chạy test ở local

Test cần toolchain local (Go/Node/Python), không chạy trong Docker:

```
make test         # chạy tất cả (Go + web/extension + ml)
make test-go      # chỉ Go
make test-web     # extension + web (jest)
make test-ml      # ml (pytest)
```

Các test tích hợp cần database sẽ tự bỏ qua khi biến `TEST_DB_URL` chưa được đặt. Để chạy chúng, trỏ `TEST_DB_URL` vào một database test riêng cho từng nhóm (deal, price, scrape, auth), vì các test này tự tạo/xóa bảng của chúng. Cách CI làm là tạo `shopass_deal_test`, `shopass_price_test`, `shopass_scrape_test` riêng - xem `.github/workflows/ci.yml`. Social login (FR-AUTH-004) có test tích hợp trong module `auth`; trỏ `TEST_DB_URL` vào một database auth riêng để chạy.

GraphQL BFF (`services/bff`) là service Node, không nằm trong `make test`. Chạy test riêng: `cd services/bff && npm install && npm test` (chạy `tsc` rồi `node --test`).

Một điểm cần biết: bài test migration của module `db` dùng testcontainers (tự bật một container `timescale/timescaledb`), nên cần Docker để chạy; và một bài test Prophet cần CmdStan. Cả hai chạy được trên CI (đã cài sẵn).

## 6. Bảng lệnh

`make help` liệt kê mọi lệnh kèm mô tả ngắn. Nhóm chính: setup (`env`), vòng đời stack (`up`, `down`, `restart`, `logs`, `ps`, `migrate`, `dev-db`, `clean`), job demo (`seed`, `scrape`, `forecast`, `deal-once`, `smoke`, `psql`), test (`test`, `test-go`, `test-web`, `test-ml`).

## 7. Xử lý sự cố

Cổng bị trùng (5432/8081/3000 đã có tiến trình khác): sửa `DB_PORT`, `PRICE_PORT`, hoặc `WEB_PORT` trong `deploy/.env` rồi `make up` lại.

Service báo lỗi kết nối DB lúc khởi động: `db` cần vài giây để sẵn sàng. Compose đã có healthcheck và `migrate` chỉ chạy sau khi `db` healthy; nếu vẫn lỗi, xem `make logs` và thử `make restart`.

Sửa code nhưng không thấy đổi: build lại image bằng `make up` (đã kèm `--build`). Nếu nghi ngờ cache, chạy `docker compose -f deploy/docker-compose.yml build --no-cache <service>`.

Muốn làm sạch hoàn toàn (kể cả dữ liệu): `make clean` rồi `make up`.

Kiến trúc và phạm vi tính năng hiện có: xem `docs/AUDIT-REPORT.md` và `docs/FR-COVERAGE.md`. Triển khai lên server: xem `deploy/README.md`.
