# SănDeal (shopass)

Nền tảng SaaS-tiện ích săn deal / theo dõi giá / tối ưu mua sắm đa sàn (Shopee + TikTok Shop + Lazada) cho Việt Nam và Đông Nam Á.

Trạng thái: đang build theo Spec Kit. 90 Feature Request + 10 NFR đặc tả toàn bộ sản phẩm, mỗi cái đã qua audit độc lập 10/10.

Layer 0/P0 foundation đã hoàn tất qua `T021`. Task tiếp theo là `T022`; batch kế tiếp là `T022-T031`, sau đó review ở `T032`.

## Bắt đầu ở đâu

Đọc hướng dẫn onboarding đầy đủ (cho cả người mới lẫn agent triển khai): [`docs/feature-requests/README.md`](docs/feature-requests/README.md).

Lối tắt tới các tài liệu chính:

- [`docs/PROJECT-STRUCTURE.md`](docs/PROJECT-STRUCTURE.md) - bản đồ repo: docs, specs, code, deploy, tests.
- [`docs/feature-requests/BACKLOG.md`](docs/feature-requests/BACKLOG.md) - index 90 FR theo phase -> module -> slice (bức tranh tổng).
- [`docs/feature-requests/SHIP-GUIDE.md`](docs/feature-requests/SHIP-GUIDE.md) - conventions build + 9 bất biến không thương lượng (đọc trước khi build).
- [`docs/feature-requests/IMPLEMENTATION-ORDER.md`](docs/feature-requests/IMPLEMENTATION-ORDER.md) - thứ tự build theo 8 layer (chọn FR tiếp theo).
- [`docs/feature-requests/DATA-MODEL.md`](docs/feature-requests/DATA-MODEL.md) - schema DB hợp nhất.
- [`docs/feature-requests/STATUS-REFERENCE.md`](docs/feature-requests/STATUS-REFERENCE.md) - enum 10 trạng thái + vòng đời FR.
- [`specs/001-full-project-plan/README.md`](specs/001-full-project-plan/README.md) - Spec Kit execution pack: plan, tasks, evidence, release gates.
- [`specs/001-full-project-plan/tasks.md`](specs/001-full-project-plan/tasks.md) - task board bắt buộc làm theo thứ tự.
- Tài liệu nền tảng (PRD + SRS + chiến lược): trong [`docs/`](docs/).

## Bố cục

```
shopass/
  README.md                          # file này (landing)
  AGENTS.md                          # giao thức memory CyberOS (BRAIN) - KHÔNG phải conventions build
  docs/
    PROJECT-STRUCTURE.md             # bản đồ repo và ranh giới docs/specs/code
    TÀI LIỆU NỀN TẢNG ... SănDeal.md  # tài liệu nguồn
    feature-requests/                # 90 FR (16 module) + BACKLOG + SHIP-GUIDE + IMPLEMENTATION-ORDER
                                     #   + DATA-MODEL + STATUS-REFERENCE + README chi tiết
    non-functional-requirements/     # 10 NFR + audit
  specs/001-full-project-plan/       # Spec Kit plan/tasks/evidence/release gates
  db/ services/ secrets/ extension/  # implementation code
  web/ mobile/ ml/ deploy/ tests/    # implementation, deploy, and validation surfaces
```

`AGENTS.md` ở gốc dành cho giao thức memory CyberOS; conventions build SănDeal nằm ở `docs/feature-requests/SHIP-GUIDE.md`.

`.agents/`, `.specify/`, `.codex/`, `node_modules/`, build outputs, `.gitkeep`, `.DS_Store`, and `.fuse_hidden*` are local/tooling noise and should not be committed.
