# Shopass

Nền tảng SaaS-tiện ích săn deal / theo dõi giá / tối ưu mua sắm đa sàn (Shopee + TikTok Shop + Lazada) cho Việt Nam và Đông Nam Á.

Trạng thái: backlog đầy đủ (90 Task + 10 NFR, mỗi cái qua audit độc lập 10/10) và một phần code đã hiện thực. Vòng lặp lõi chạy được end to end: scrape giá Shopee, lưu vào TimescaleDB, dự báo đáy, bắn cảnh báo, hiển thị biểu đồ. Xem `docs/AUDIT-REPORT.md` (chất lượng, cách chạy test) và `docs/TASK-COVERAGE.md` (phần nào thật, phần nào còn stub).

## Chạy nhanh bằng Docker

```
make env      # tạo deploy/.env
make up       # dựng cả stack (db + service)
make smoke    # đi hết vòng lặp lõi và in kết quả
```

Kỹ sư dev local: [`docs/DEVELOPMENT-GUIDE.md`](docs/DEVELOPMENT-GUIDE.md). Triển khai server (DevOps): [`deploy/README.md`](deploy/README.md). Danh sách lệnh: `make help`.

## Bắt đầu ở đâu

Đọc hướng dẫn onboarding đầy đủ (cho cả người mới lẫn agent triển khai): [`docs/tasks/README.md`](docs/tasks/README.md).

Lối tắt tới các tài liệu chính:

- [`docs/tasks/BACKLOG.md`](docs/tasks/BACKLOG.md) - index 90 task theo phase -> module -> slice (bức tranh tổng).
- [`docs/tasks/SHIP-GUIDE.md`](docs/tasks/SHIP-GUIDE.md) - conventions build + 9 bất biến không thương lượng (đọc trước khi build).
- [`docs/tasks/IMPLEMENTATION-ORDER.md`](docs/tasks/IMPLEMENTATION-ORDER.md) - thứ tự build theo 8 layer (chọn task tiếp theo).
- [`docs/tasks/DATA-MODEL.md`](docs/tasks/DATA-MODEL.md) - schema DB hợp nhất.
- [`docs/tasks/STATUS-REFERENCE.md`](docs/tasks/STATUS-REFERENCE.md) - enum 10 trạng thái + vòng đời task.
- [`docs/tasks/ANTIGRAVITY-KICKOFF.md`](docs/tasks/ANTIGRAVITY-KICKOFF.md) - prompt dán vào Antigravity để bắt đầu build.
- Tài liệu nền tảng (PRD + SRS + chiến lược): trong [`docs/`](docs/).

## Bố cục

```
shopass/
  README.md                          # file này (landing)
  AGENTS.md                          # giao thức memory CyberOS (BRAIN) - KHÔNG phải conventions build
  docs/
    TÀI LIỆU NỀN TẢNG ... (PRD/SRS).md  # tài liệu nguồn (tên file lịch sử có thể còn SănDeal)
    tasks/                # 90 task (16 module) + BACKLOG + SHIP-GUIDE + IMPLEMENTATION-ORDER
                                     #   + DATA-MODEL + STATUS-REFERENCE + README chi tiết
    non-functional-requirements/     # 10 NFR + audit
  services/ extension/ web/          # code hiện có
  # mobile/                          # chưa có trong repo (P3; xem docs/TASK-COVERAGE.md)
```

`AGENTS.md` ở gốc dành cho giao thức memory CyberOS; conventions build Shopass nằm ở `docs/tasks/SHIP-GUIDE.md`. Tài liệu nền tảng lịch sử vẫn có thể dùng tên SănDeal; thương hiệu sản phẩm user-facing là Shopass.
