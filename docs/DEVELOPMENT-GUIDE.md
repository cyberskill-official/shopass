# SănDeal (shopass) - Hướng Dẫn Phát Triển Local

Tài liệu này cung cấp hướng dẫn từng bước để các kỹ sư thiết lập và chạy nền tảng SănDeal trong môi trường phát triển local (máy cá nhân).

## 1. Yêu cầu hệ thống (Prerequisites)

Trước khi bắt đầu, hãy đảm bảo bạn đã cài đặt các công cụ sau trên máy:
- **Go** (1.22+)
- **Python** (3.12+) & **uv** (dành cho service ML)
- **Node.js** (18+) & **npm** (dành cho web frontend)
- **PostgreSQL** (15+)
- **Playwright** (dành cho scraper service)
- **Git**

## 2. Thiết lập Cơ sở dữ liệu (PostgreSQL)

SănDeal sử dụng một database PostgreSQL hợp nhất cho tất cả các services.

1. **Khởi động PostgreSQL** trên máy local (thông qua Docker, Postgres.app, hoặc homebrew).
2. **Tạo database**:
   ```bash
   createdb shopass_db
   ```
3. **Chạy Migrations**:
   Các file migration được đặt trong thư mục `db/migrations/`. Bạn có thể áp dụng chúng bằng công cụ migration tùy thích (vd: `golang-migrate`) hoặc chạy thủ công các file SQL theo thứ tự:
   ```bash
   # Ví dụ sử dụng golang-migrate
   migrate -path db/migrations -database "postgres://localhost:5432/shopass_db?sslmode=disable" up
   ```
4. **Seed Database (Tùy chọn nhưng khuyến nghị)**:
   Chạy các dữ liệu mẫu (seed) trong `db/seed/` để khởi tạo các nền tảng (platforms), policies ban đầu, v.v.

## 3. Backend Services (Go)

Backend bao gồm nhiều Go microservices nằm trong thư mục `services/`.

### Biến môi trường (Environment Variables)
Khi phát triển ở local, thông thường bạn sẽ cần export các biến môi trường sau (có thể tạo file `.env` hoặc export trực tiếp trên terminal):
```bash
export DATABASE_URL="postgres://localhost:5432/shopass_db?sslmode=disable"
export JWT_SECRET="your-local-dev-jwt-secret"
export CGO_ENABLED=0 # Khuyến nghị cho macOS
```

### Chạy các Services Cốt lõi
Mỗi service có thể chạy độc lập. Mở nhiều tab terminal hoặc dùng một công cụ multiplexer (như `tmux`) để chạy các services cần thiết:

**1. API Gateway** (Định tuyến requests & xác thực JWTs):
```bash
go run services/gateway/cmd/gateway/main.go
```
*Lưu ý: Đảm bảo port của gateway (mặc định thường là 8080) không bị trùng.*

**2. Auth Service** (Xử lý đăng nhập/đăng ký):
```bash
go run services/auth/cmd/auth/main.go
```

**3. Deal Service** (Quản lý deal & HTTP fanout tới Notifsvc):
```bash
go run services/deal/cmd/dealsvc/main.go
```

**4. Notification Service** (Xử lý gửi email/push notifications):
```bash
go run services/notif/cmd/notifsvc/main.go
```

**5. Scrape Service** (Sử dụng local Playwright để cào dữ liệu TikTok/Shopee):
```bash
# Đảm bảo đã cài đặt trình duyệt playwright trước!
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.2000.1 install --with-deps

go run services/scrape/cmd/scrape/main.go
```

*(Bạn có thể chạy các services khác như `bill`, `cart`, `comply`, `affil`, `track`, `price` tương tự khi cần cho các task cụ thể).*

## 4. Machine Learning Service (Python)

Service ML xử lý dự đoán giá và các tác vụ NLP.

1. Di chuyển vào thư mục ML:
   ```bash
   cd services/ml
   ```
2. Tạo và kích hoạt môi trường ảo (virtual environment) bằng `uv`:
   ```bash
   uv venv
   source .venv/bin/activate
   ```
3. Cài đặt các thư viện phụ thuộc:
   ```bash
   uv pip install -r requirements.txt
   ```
4. Chạy ML API (giả sử dùng một runner tiêu chuẩn như uvicorn):
   ```bash
   python main.py # hoặc uvicorn main:app --reload
   ```

## 5. Web Frontend (React/Next.js/Vite)

Ứng dụng web được đặt trong thư mục `web/`.

1. Di chuyển vào thư mục web:
   ```bash
   cd web
   ```
2. Cài đặt dependencies:
   ```bash
   npm install
   ```
3. Khởi động development server:
   ```bash
   npm run dev
   ```
4. Mở trình duyệt và truy cập `http://localhost:3000` (hoặc port được chỉ định trên console).

## 6. Chrome Extension

Để test Chrome extension của SănDeal ở môi trường local:

1. Mở Google Chrome.
2. Truy cập vào địa chỉ `chrome://extensions/`.
3. Bật chế độ **Developer mode** ở góc trên cùng bên phải.
4. Bấm vào **Load unpacked**.
5. Chọn thư mục `extension/` trong repo `shopass`.

Mỗi khi bạn có thay đổi file trong thư mục extension, bạn có thể cần phải bấm vào icon "Refresh" trên thẻ extension ở trang `chrome://extensions/`.

---

### Khắc phục sự cố (Troubleshooting)

- **Lỗi CGO/SQLite**: Hãy chắc chắn đã set `CGO_ENABLED=0` nếu bạn gặp lỗi biên dịch binary trên macOS, ngoại trừ trường hợp một module cụ thể (như `go-sqlite3`) bắt buộc cần CGO. (Dự án dùng Postgres, nên có thể yên tâm tắt CGO).
- **Lỗi Playwright Mismatch**: Đảm bảo phiên bản Go package của Playwright trong `scrape` khớp với các trình duyệt đã cài đặt (`v0.2000.1`).
- **Thiếu biến môi trường**: Kiểm tra log của các services bị lỗi. Nếu một service bị panic khi khởi động, rất có thể là do thiếu `DATABASE_URL` hoặc port cấu hình.
