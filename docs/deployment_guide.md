# SănDeal - Hướng dẫn Triển khai & Phát hành Toàn diện (Deployment Guide)

> **Trạng thái:** tài liệu này là bản hướng dẫn kiến trúc/rộng. Quy trình Docker Compose an toàn hiện tại nằm ở [`../deploy/README.md`](../deploy/README.md). Không mở công khai `deploy/docker-compose.yml`, không dùng `Caddyfile.demo`, và không xem các lệnh trong file này là checklist production đã đủ. `deploy/README.md` nêu rõ các release blocker còn lại: gateway JWT thật, khoá ký auth bền vững, secret loading, readiness, backup/restore và notification.

Tài liệu này cung cấp các bước chi tiết (step-by-step) để bạn có thể tự tay (manually) triển khai toàn bộ hệ sinh thái SănDeal lên môi trường Production, bao gồm Backend (Go), ML (Python), Web Frontend (Next.js), và xuất bản Extension lên Chrome Web Store.

---

## Phần 1: Chuẩn bị Hạ tầng (Infrastructure)

Hệ thống yêu cầu các thành phần nền tảng sau:

1. **PostgreSQL 16 + TimescaleDB**: Dùng cho `app_user`, `price_snapshot`, `wishlist`, `cart_snapshot`, v.v. (Nên dùng AWS RDS hoặc dịch vụ managed DB tương tự để dễ cấu hình TimescaleDB).
2. **Redis & Kafka (hoặc Redis Streams)**: Làm Message Queue cho hệ thống Fan-out notification (TASK-NOTIF-003) và caching.
3. **HashiCorp Vault / AWS Secrets Manager**: Dùng để quản lý secrets (TASK-INFRA-003). KHÔNG được lưu cleartext mật khẩu trong mã nguồn hay biến môi trường thông thường.

### Bước 1.1: Chạy Migrations (Khởi tạo Database)

Tất cả các service (ví dụ `services/bill`, `services/price`, `services/auth`, v.v.) đều có thư mục `migrations/`. Bạn cần áp dụng các file SQL này vào cơ sở dữ liệu:

```bash
# Ví dụ cấu hình cho service Auth
export DB_URL="postgres://user:pass@host:5432/sandeal_prod?sslmode=require"

# Dùng công cụ golang-migrate hoặc apply thủ công
for svc in auth bill cart comply deal ext gateway infra notif price scrape track trust; do
  psql $DB_URL -f services/$svc/migrations/*.sql
done
```

### Bước 1.2: Thiết lập Vault (Secrets)

Lưu trữ các key sau vào Vault của bạn (để các backend service lấy lúc khởi động):
- JWT Signing Key (`JWT_SECRET`)
- Database Connection String
- Các API Keys của bên thứ 3 (FCM/APNs, Email SMTP, SMS gateway, Playwright Proxy auth, Involve Asia/Accesstrade token).

---

## Phần 2: Triển khai Backend (Go Services)

Các dịch vụ Go được xây dựng độc lập trong thư mục `services/`.

### Bước 2.1: Biên dịch (Build)
Chạy trên máy chủ hoặc trong quy trình CI/CD:

```bash
# Đứng tại thư mục gốc của project
cd services/gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /deploy/gateway ./cmd/gateway

cd ../bill
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /deploy/bill ./cmd/bill

# Lặp lại với các service khác (auth, scrape, price, v.v...)
```

### Bước 2.2: Chạy Service (Run)
Có thể sử dụng systemd, Docker, hoặc Kubernetes. Dưới đây là ví dụ dùng PM2 hoặc lệnh thô:

```bash
# Cung cấp biến môi trường trỏ tới Vault
export VAULT_ADDR="https://vault.sandeal.io"
export VAULT_TOKEN="<token>"

# Khởi chạy API Gateway (BFF)
./deploy/gateway &

# Khởi chạy các worker / service
./deploy/bill &
./deploy/price &
# ...
```

---

## Phần 3: Triển khai Machine Learning (Python)

ML module nằm tại `services/ml`, sử dụng `uv` làm package manager. Đảm nhiệm việc chạy batch scoring vào ban đêm để dự đoán đáy giá (LightGBM/Prophet).

### Bước 3.1: Môi trường
```bash
cd services/ml
uv sync # Cài đặt dependencies từ requirements.txt / pyproject.toml
```

### Bước 3.2: Cronjob (Batch Scoring)
Cấu hình Cronjob (hoặc Apache Airflow) chạy vào 00:00 mỗi đêm (TASK-DEAL-006):

```bash
# Trong crontab -e của server:
0 0 * * * cd /path/to/sandeal/services/ml && PYTHONPATH=. uv run python scripts/batch_score_nightly.py >> /var/log/sandeal_ml.log 2>&1
```

---

## Phần 4: Triển khai Web Frontend (Next.js)

Thư mục `web/` chứa giao diện web tĩnh (SEO landing page) và web app (bảng điều khiển, biểu đồ giá).

### Bước 4.1: Build Next.js
```bash
cd web
npm install
npm run build
```

### Bước 4.2: Start Next.js (Node.js Server)
```bash
npm run start
```
*Ghi chú: Nếu bạn host lên Vercel, chỉ cần kết nối repo với Vercel và cấu hình Root Directory là `web`.*

---

## Phần 5: Đóng gói & Phát hành Chrome Extension

Đây là bước cực kỳ quan trọng vì extension (MV3) liên quan trực tiếp đến compliance (PDPL và chính sách Chrome Web Store). Extenstion nằm tại thư mục `extension/`.

### Bước 5.1: Biên dịch Extension
Bạn có thể tự build file `.zip` bằng command:

```bash
cd extension
npm install
npm run build
# Script build sẽ tạo ra thư mục dist/ hoặc file sandeal-extension.zip
```
*(Nếu sử dụng kịch bản zip cơ bản: `zip -r sandeal-ext.zip dist/`)*

### Bước 5.2: Đăng ký Chrome Web Store
1. Truy cập [Chrome Developer Dashboard](https://chrome.google.com/webstore/devconsole/).
2. Đăng nhập / Đăng ký tài khoản nhà phát triển (phí $5 trọn đời).
3. Bấm **"New Item"** và upload file `sandeal-ext.zip` (hoặc `.zip` của thư mục `dist`).

### Bước 5.3: Hoàn thành Khai báo Bảo mật & Riêng tư (BẮT BUỘC ĐỂ DUYỆT)
Để extension không bị gỡ (như rủi ro ghi nhận ở BACKLOG), bạn **PHẢI** khai báo trung thực:

1. **Single Purpose:** "Extension giúp người dùng theo dõi giá trị thực của sản phẩm, chống sale ảo và tối ưu hóa giỏ hàng trên các sàn TMĐT."
2. **Permissions Justification (Lý do cấp quyền):**
- `host_permissions` (`*://*.shopee.vn/*`, `*://*.tiktok.com/*`...): Bắt buộc để đọc thông tin giá và vouchers từ DOM.
- `storage` / `alarms`: Cần thiết cho Manifest V3 Service Worker để lưu cache và chạy background checks.
- `declarativeNetRequest`: Dùng để ngăn chặn hoặc định tuyến các tracking requests dư thừa (bảo vệ quyền riêng tư người dùng).
3. **Data Usage (Sử dụng dữ liệu):**
- Extension **KHÔNG GỬI** cookie, session token, hay mật khẩu rời khỏi máy người dùng. (Điều này đã được chứng minh qua file `audit/run-security-audit.sh`).
- Tích chọn: *"This item only uses data for its core functionality"*.
- Khẳng định KHÔNG bán dữ liệu cá nhân cho bên thứ ba.
4. **Affiliate Disclosure (Khai báo Tiếp thị liên kết):**
- Trong mô tả Cửa hàng, **PHẢI GHI RÕ**: "SănDeal có thể nhận được hoa hồng khi bạn mua sắm qua các liên kết (hoàn toàn do người dùng chủ động click tạo link)". Không dùng auto-cookie-stuffing (chính sách TASK-AFFIL-004 đã enforce ở code, nhưng bạn phải ghi rõ trên text).

### Bước 5.4: Nộp duyệt (Publish)
Nhấn **Submit for Review**. Quá trình review có thể mất từ 1-3 ngày cho lần đầu tiên. Do extension có host permissions rộng, Google có thể review thủ công khá kỹ.

---

## Phần 6: Tuân thủ Pháp lý PDPL (Luật 91/2025/QH15)

Trước khi đón user thật, bạn cần đảm bảo các hành động pháp lý thủ công (ngoài code):
1. **DPIA (Đánh giá Tác động Xử lý Dữ liệu Cá nhân):** Lập hồ sơ DPIA (theo mẫu của Bộ Công An / Cơ quan Bảo vệ Dữ liệu) và nộp trong vòng 60 ngày kể từ khi hệ thống bắt đầu xử lý dữ liệu người dùng (TASK-COMPLY-002).
2. **Quy định 72h:** Chuẩn bị sẵn một kịch bản/quy trình báo cáo khi phát hiện lộ lọt dữ liệu (TASK-COMPLY-004).

---
*Chúc bạn triển khai SănDeal thành công! Các bài toán khó nhất về Data model, Anti-bot, Affiliate Compliance và Cấu trúc mã đã được giải quyết triệt để trong source code.*
