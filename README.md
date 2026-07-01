# SănDeal (shopass)

Nền tảng SaaS-tiện ích săn deal / theo dõi giá / tối ưu mua sắm đa sàn (Shopee + TikTok Shop + Lazada) cho Việt Nam và Đông Nam Á.

Trạng thái: backlog SHIP-READY. 90 Feature Request + 10 NFR đặc tả toàn bộ sản phẩm, mỗi cái đã qua audit độc lập 10/10. Đây là spec để xây - chưa có code; mục tiêu là một người hoặc một agent đọc xong là build được.

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
