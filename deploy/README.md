# SănDeal - Hướng dẫn triển khai (DevOps)

Tài liệu này hướng dẫn triển khai SănDeal bằng Docker Compose, từng bước, cho môi trường staging/production. Stack chạy vòng lặp lõi trên một Postgres/TimescaleDB thật: scrape giá, lưu, dự báo đáy, bắn cảnh báo, hiển thị biểu đồ.

Đối tượng: DevOps/SRE. Kỹ sư dev local xem `docs/DEVELOPMENT-GUIDE.md`.

## 1. Yêu cầu trên server

- Docker Engine (khuyến nghị 24+) kèm plugin Compose v2. Kiểm tra: `docker compose version`.
- Git và make (hoặc chạy trực tiếp các lệnh `docker compose`).
- Tài nguyên tối thiểu gợi ý: 2 vCPU, 4 GB RAM, 20 GB đĩa (Postgres/TimescaleDB, image, dữ liệu giá). Dữ liệu giá tích lũy theo thời gian nên theo dõi dung lượng.

## 2. Các thành phần trong stack

Luôn bật: `db` (Postgres 16 + TimescaleDB), `migrate` (chạy 1 lần rồi thoát), `pricesvc` (API giá + so sánh chéo sàn, cổng 8081), `dealsvc` (cron chấm điểm đáy 02:00 Asia/Ho_Chi_Minh + phục vụ biểu đồ TASK-DEAL-003 ở cổng 8082), `notifsvc` (nhận thông báo), `web` (giao diện, cổng 3000), `authsvc` (đăng nhập/refresh + social login, cổng 8084), `tracksvc` (wishlist, cổng 8083), `bff` (GraphQL cho web, cổng 8085).

Job chạy theo yêu cầu (profile `jobs`): `scrapesvc` (scrape + hàng đợi bền vững), `mlforecast` (dự báo, có cài CmdStan để chạy Prophet).

Tất cả service dùng chung một database `shopass`; mỗi bảng có đúng một service sở hữu.

## 3. Cấu hình

```
git clone <repo-url> shopass
cd shopass
make env                 # tạo deploy/.env từ mẫu
```

Mở `deploy/.env` và đặt giá trị thật trước khi lên staging/production:

- `POSTGRES_PASSWORD`: bắt buộc đổi khỏi giá trị mặc định.
- `POSTGRES_USER`, `POSTGRES_DB`: đổi nếu cần.
- `DB_PORT`, `PRICE_PORT`, `WEB_PORT`: cổng mở ra host; đổi nếu trùng.
- `SHOPEE_BASE_URL`, `SCRAPE_SEED`, `HTTPS_PROXY`: cấu hình cho job scrape (proxy dân cư theo TASK-SCRAPE-002).

File `deploy/.env` đã nằm trong `.gitignore` - không commit. `docker compose` tự nạp file này.

## 4. Triển khai

```
make up                  # build image + chạy toàn bộ service luôn-bật
make ps                  # kiểm tra: db healthy, migrate đã Exit 0, còn lại Up
```

`migrate` chạy đầy đủ chuỗi migration (nền tảng dùng chung + của từng service) lên TimescaleDB thật, gồm cả hypertable `price_snapshot` và continuous aggregate `price_daily`.

## 5. Kiểm tra sau triển khai

```
curl -s http://<host>:8081/v1/products/1/price-history?range=7d   # pricesvc trả JSON
open http://<host>:3000                                            # web
make smoke                                                         # đi hết vòng lặp và in kết quả
```

`make smoke` seed một sản phẩm demo, chạy scrape + deal, rồi in `price_snapshot` và `bottom_alert_log`. Với server không ra được Shopee thật, trỏ `SHOPEE_BASE_URL` sang endpoint fixture để chạy khô.

## 6. Vận hành

Log và trạng thái:

```
make logs                # hoặc: docker compose -f deploy/docker-compose.yml logs -f <service>
make ps
```

Chạy lại migration (idempotent, sau khi cập nhật schema): `make migrate`.

Mở psql: `make psql`.

Sao lưu và phục hồi database:

```
# backup
docker compose -f deploy/docker-compose.yml exec -T db pg_dump -U postgres shopass | gzip > backup_$(date +%F).sql.gz
# restore (stack đang chạy, database rỗng)
gunzip -c backup_YYYY-MM-DD.sql.gz | docker compose -f deploy/docker-compose.yml exec -T db psql -U postgres -d shopass
```

Giám sát: thư mục `deploy/grafana/` có sẵn dashboard. Đấu Prometheus/Grafana vào endpoint metrics của các service khi bạn thêm chúng vào compose (chưa nằm trong stack lõi này).

Lên lịch job (quan trọng để hệ tự chạy): `scrapesvc` và `mlforecast` là job một-lần. Đặt cron trên host, hoặc thêm một scheduler (ví dụ ofelia) để chạy định kỳ:

```
# ví dụ crontab trên host: scrape mỗi 5 phút, dự báo mỗi giờ
*/5 * * * * cd /srv/shopass && SCRAPE_SEED= docker compose -f deploy/docker-compose.yml run --rm scrapesvc
0   * * * * cd /srv/shopass && docker compose -f deploy/docker-compose.yml run --rm mlforecast
```

Hàng đợi `scrape_job` là bền vững trên Postgres, nên chạy chồng hay service chết giữa chừng đều an toàn (lease + reclaim).

Mở rộng: tăng số bản sao service không giữ trạng thái:

```
docker compose -f deploy/docker-compose.yml up -d --scale pricesvc=3
```

Cập nhật phiên bản:

```
git pull
make up                  # build lại image thay đổi và chạy tiếp; make migrate nếu có migration mới
```

Dừng và dọn:

```
make down                # dừng, GIỮ dữ liệu
make clean               # dừng và XÓA volume dữ liệu (mất toàn bộ dữ liệu)
```

## 7. Bảo mật

- Đổi `POSTGRES_PASSWORD` và mọi bí mật khỏi giá trị mặc định.
- Không mở cổng `db` (5432) ra Internet công cộng; chỉ mở cho mạng nội bộ hoặc bỏ ánh xạ cổng nếu không cần truy cập từ host.
- Đặt một reverse proxy có TLS (Caddy/Nginx/Traefik) trước `web` và `pricesvc`; không phục vụ HTTP trần ra ngoài.
- Bí mật thật (khóa cổng thanh toán, token proxy) nên đưa qua trình quản lý bí mật hoặc biến môi trường của môi trường triển khai, không ghi vào file trong repo.
- Tôn trọng các bất biến ở `docs/tasks/SHIP-GUIDE.md` (không lưu token phiên sàn phía server, tuân thủ PDPL).

## 8. Ghi chú và giới hạn hiện tại

- Ảnh Go build dùng `golang:1.25` (khớp `go.mod`); nếu hạ về 1.22 theo ghi chú vệ sinh trong `docs/AUDIT-REPORT.md` thì đổi cả ở đây.
- Ảnh ml cài CmdStan để chạy Prophet cho SKU đã đủ lịch sử; SKU chưa đủ dùng đường prior nhẹ.
- Chưa nằm trong stack: API gateway đứng trước các service (xác thực JWT tập trung). Hiện `authsvc`, `tracksvc`, `bff` tin `X-User-Id` do gateway chuyển xuống; khi chưa có gateway thật, đặt gateway/reverse-proxy tự gắn header này, hoặc chỉ mở các service qua mạng nội bộ.
- `authsvc`, `tracksvc`, `bff` nay đã nằm trong compose và build được (đã kiểm bằng `go build`/`npm run build`), nhưng `docker compose up` chưa chạy thử ở môi trường xây dựng vì không có Docker. Còn hai điểm chưa trọn: `bff` trả wishlist nhưng chưa kèm item (cần `tracksvc` nhúng item vào response `GET /v1/wishlists`), và social login chỉ bật khi có `GOOGLE_CLIENT_ID/SECRET` thật.
- Chuỗi migration `deploy/migrate.sh` (gồm cả `0007_social_identity`) đã được kiểm trên một Postgres thật và áp dụng sạch; `make migrate` tạo cả `social_identity`.
- Đọc thêm: `docs/AUDIT-REPORT.md` (chất lượng và cách chạy test), `docs/TASK-COVERAGE.md` (tính năng thật và phần còn stub).
