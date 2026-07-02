# SănDeal (shopass)

Nền tảng SaaS-tiện ích săn deal / theo dõi giá / tối ưu mua sắm đa sàn (Shopee + TikTok Shop + Lazada) cho Việt Nam và Đông Nam Á.

Trạng thái: backlog đầy đủ (90 Feature Request + 10 NFR, mỗi cái qua audit độc lập 10/10) và một phần code đã hiện thực. Vòng lặp lõi chạy được end to end: scrape giá Shopee, lưu vào TimescaleDB, dự báo đáy, bắn cảnh báo, hiển thị biểu đồ. Xem `docs/AUDIT-REPORT.md` (chất lượng, cách chạy test) và `docs/FR-COVERAGE.md` (phần nào thật, phần nào còn stub).

## Chạy nhanh bằng Docker

```
make env      # tạo deploy/.env
make up       # dựng cả stack (db + service)
make smoke    # đi hết vòng lặp lõi và in kết quả
```

Kỹ sư dev local: [`docs/DEVELOPMENT-GUIDE.md`](docs/DEVELOPMENT-GUIDE.md). Triển khai server (DevOps): [`deploy/README.md`](deploy/README.md). Danh sách lệnh: `make help`.

## Bắt đầu ở đâu

Đọc hướng dẫn onboarding đầy đủ (cho cả người mới lẫn agent triển khai): [`docs/feature-requests/README.md`](docs/feature-requests/README.md).

Lối tắt tới các tài liệu chính:

- [`docs/feature-requests/BACKLOG.md`](docs/feature-requests/BACKLOG.md) - index 90 FR theo phase -> module -> slice (bức tranh tổng).
- [`docs/feature-requests/SHIP-GUIDE.md`](docs/feature-requests/SHIP-GUIDE.md) - conventions build + 9 bất biến không thương lượng (đọc trước khi build).
- [`docs/feature-requests/IMPLEMENTATION-ORDER.md`](docs/feature-requests/IMPLEMENTATION-ORDER.md) - thứ tự build theo 8 layer (chọn FR tiếp theo).
- [`docs/feature-requests/DATA-MODEL.md`](docs/feature-requests/DATA-MODEL.md) - schema DB hợp nhất.
- [`docs/feature-requests/STATUS-REFERENCE.md`](docs/feature-requests/STATUS-REFERENCE.md) - enum 10 trạng thái + vòng đời FR.
- [`docs/feature-requests/ANTIGRAVITY-KICKOFF.md`](docs/feature-requests/ANTIGRAVITY-KICKOFF.md) - prompt dán vào Antigravity để bắt đầu build.
- Tài liệu nền tảng (PRD + SRS + chiến lược): trong [`docs/`](docs/).

## Bố cục

```
shopass/
  README.md                          # file này (landing)
  AGENTS.md                          # giao thức memory CyberOS (BRAIN) - KHÔNG phải conventions build
  docs/
    TÀI LIỆU NỀN TẢNG ... SănDeal.md  # tài liệu nguồn
    feature-requests/                # 90 FR (16 module) + BACKLOG + SHIP-GUIDE + IMPLEMENTATION-ORDER
                                     #   + DATA-MODEL + STATUS-REFERENCE + README chi tiết
    non-functional-requirements/     # 10 NFR + audit
  services/ extension/ web/ mobile/  # (sẽ được agent tạo khi build theo new_files của từng FR)
```

`AGENTS.md` ở gốc dành cho giao thức memory CyberOS; conventions build SănDeal nằm ở `docs/feature-requests/SHIP-GUIDE.md`.
