# SănDeal - Thứ tự triển khai (build order)

Tài liệu này dẫn xuất tự động từ đồ thị `depends_on` của 90 task. Một agent triển khai nên build theo **layer**: mọi task trong layer N chỉ phụ thuộc vào task ở layer < N, nên trong cùng một layer các task build song song được. DAG đã được xác minh **acyclic** (không có phụ thuộc vòng) và **reciprocal** (`blocks` là nghịch đảo của `depends_on`).

Tổng: **90 task**, **627h**, **8 layer**. Mỗi task ở status `done` (audit 10/10).

## Cách dùng cho agent ship

1. Lấy task khả thi tiếp theo = task ở layer thấp nhất chưa `done` mà mọi `depends_on` đã `done`.
2. Đọc file task (frontmatter + §1-§11), build theo `new_files`/`sub_tasks`, chạy test §5, đối chiếu acceptance criteria §4.
3. Cập nhật status theo `STATUS-REFERENCE.md` (`implementing` -> `ready_to_review` -> ... -> `done`).
4. Ưu tiên trong cùng layer: `MUST` trước `SHOULD` trước `COULD`.

## Layer 0 - 4 task, 27h

| task | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| TASK-EXT-001 | MUST | P1 | ext | - | 8 |
| TASK-INFRA-001 | MUST | P0 | infra | - | 8 |
| TASK-INFRA-002 | MUST | P0 | infra | - | 6 |
| TASK-INFRA-003 | MUST | P0 | infra | - | 5 |

## Layer 1 - 12 task, 78h

| task | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| TASK-AFFIL-001 | MUST | P2 | affil | TASK-INFRA-002 | 6 |
| TASK-AUTH-001 | MUST | P1 | auth | TASK-INFRA-002 | 6 |
| TASK-CART-001 | MUST | P2 | cart | TASK-INFRA-002 | 6 |
| TASK-COMPLY-001 | MUST | P1 | comply | TASK-INFRA-002 | 8 |
| TASK-COMPLY-005 | MUST | P1 | comply | TASK-INFRA-003 | 5 |
| TASK-EXT-002 | MUST | P1 | ext | TASK-EXT-001 | 10 |
| TASK-INFRA-004 | MUST | P0 | infra | TASK-INFRA-001 | 8 |
| TASK-INFRA-005 | MUST | P0 | infra | TASK-INFRA-002 | 6 |
| TASK-NOTIF-001 | MUST | P1 | notif | TASK-INFRA-002 | 6 |
| TASK-PRICE-001 | MUST | P1 | price | TASK-INFRA-002 | 6 |
| TASK-TRUST-001 | MUST | P1 | trust | TASK-EXT-001 | 6 |
| TASK-EXT-004 | SHOULD | P1 | ext | TASK-EXT-001 | 5 |

## Layer 2 - 19 task, 130h

| task | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| TASK-AFFIL-002 | MUST | P2 | affil | TASK-AFFIL-001 | 6 |
| TASK-AUTH-002 | MUST | P1 | auth | TASK-AUTH-001, TASK-INFRA-001 | 6 |
| TASK-AUTH-003 | MUST | P1 | auth | TASK-AUTH-001 | 5 |
| TASK-BILL-001 | MUST | P2 | bill | TASK-AUTH-001 | 6 |
| TASK-CART-005 | MUST | P2 | cart | TASK-EXT-002, TASK-CART-001 | 6 |
| TASK-COMPLY-002 | MUST | P1 | comply | TASK-COMPLY-001 | 6 |
| TASK-COMPLY-003 | MUST | P1 | comply | TASK-COMPLY-001 | 8 |
| TASK-COMPLY-004 | MUST | P1 | comply | TASK-COMPLY-001, TASK-INFRA-004 | 5 |
| TASK-EXT-003 | MUST | P1 | ext | TASK-EXT-002 | 6 |
| TASK-EXT-006 | MUST | P1 | ext | TASK-EXT-001, TASK-COMPLY-001 | 5 |
| TASK-EXT-007 | MUST | P2 | ext | TASK-EXT-002 | 10 |
| TASK-EXT-008 | MUST | P2 | ext | TASK-EXT-002 | 8 |
| TASK-NOTIF-002 | MUST | P1 | notif | TASK-NOTIF-001 | 8 |
| TASK-NOTIF-003 | MUST | P1 | notif | TASK-NOTIF-001 | 8 |
| TASK-PRICE-002 | MUST | P1 | price | TASK-PRICE-001 | 8 |
| TASK-PRICE-005 | MUST | P1 | price | TASK-PRICE-001 | 8 |
| TASK-SCRAPE-001 | MUST | P1 | scrape | TASK-INFRA-003, TASK-PRICE-001 | 10 |
| TASK-CART-006 | SHOULD | P2 | cart | TASK-EXT-002 | 5 |
| TASK-COMPLY-008 | SHOULD | P3 | comply | TASK-COMPLY-001 | 6 |

## Layer 3 - 24 task, 162h

| task | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| TASK-AFFIL-003 | MUST | P2 | affil | TASK-AFFIL-002 | 8 |
| TASK-AFFIL-004 | MUST | P2 | affil | TASK-AFFIL-002, TASK-EXT-003 | 5 |
| TASK-AUTH-005 | MUST | P1 | auth | TASK-AUTH-001, TASK-COMPLY-003 | 5 |
| TASK-BILL-002 | MUST | P2 | bill | TASK-BILL-001 | 10 |
| TASK-CART-002 | MUST | P2 | cart | TASK-EXT-003 | 5 |
| TASK-DEAL-001 | MUST | P1 | deal | TASK-PRICE-002 | 8 |
| TASK-DEAL-003 | MUST | P1 | deal | TASK-PRICE-002 | 5 |
| TASK-EXT-005 | MUST | P1 | ext | TASK-EXT-003, TASK-AUTH-002 | 6 |
| TASK-NOTIF-004 | MUST | P1 | notif | TASK-NOTIF-003 | 6 |
| TASK-NOTIF-005 | MUST | P2 | notif | TASK-NOTIF-003 | 5 |
| TASK-NOTIF-006 | MUST | P2 | notif | TASK-NOTIF-003 | 4 |
| TASK-PRICE-003 | MUST | P1 | price | TASK-PRICE-002 | 5 |
| TASK-PRICE-004 | MUST | P1 | price | TASK-PRICE-005 | 6 |
| TASK-SCRAPE-002 | MUST | P1 | scrape | TASK-SCRAPE-001 | 8 |
| TASK-SCRAPE-003 | MUST | P1 | scrape | TASK-SCRAPE-001 | 12 |
| TASK-TRUST-002 | MUST | P1 | trust | TASK-EXT-003 | 5 |
| TASK-TRUST-003 | MUST | P1 | trust | TASK-EXT-003, TASK-COMPLY-005 | 6 |
| TASK-WEB-001 | MUST | P1 | web | TASK-AUTH-002 | 8 |
| TASK-AUTH-004 | SHOULD | P1 | auth | TASK-AUTH-002 | 6 |
| TASK-B2B-001 | SHOULD | P3 | b2b | TASK-PRICE-002, TASK-COMPLY-003 | 10 |
| TASK-BILL-004 | SHOULD | P2 | bill | TASK-BILL-001 | 5 |
| TASK-BILL-005 | SHOULD | P2 | bill | TASK-BILL-001 | 6 |
| TASK-MOBILE-001 | SHOULD | P3 | mobile | TASK-AUTH-002, TASK-NOTIF-002 | 12 |
| TASK-NOTIF-007 | SHOULD | P2 | notif | TASK-NOTIF-003 | 6 |

## Layer 4 - 17 task, 125h

| task | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| TASK-BILL-003 | MUST | P2 | bill | TASK-BILL-002 | 6 |
| TASK-CART-003 | MUST | P2 | cart | TASK-CART-001, TASK-CART-002 | 12 |
| TASK-DEAL-002 | MUST | P1 | deal | TASK-DEAL-001 | 6 |
| TASK-SCRAPE-004 | MUST | P1 | scrape | TASK-SCRAPE-003 | 8 |
| TASK-SCRAPE-005 | MUST | P1 | scrape | TASK-SCRAPE-002, TASK-PRICE-002 | 6 |
| TASK-SCRAPE-006 | MUST | P1 | scrape | TASK-SCRAPE-002 | 6 |
| TASK-SCRAPE-007 | MUST | P2 | scrape | TASK-SCRAPE-003 | 10 |
| TASK-SCRAPE-008 | MUST | P2 | scrape | TASK-SCRAPE-003 | 8 |
| TASK-TRACK-001 | MUST | P1 | track | TASK-PRICE-001, TASK-SCRAPE-002 | 5 |
| TASK-TRUST-004 | MUST | P3 | trust | TASK-BILL-004, TASK-AFFIL-001 | 10 |
| TASK-WEB-002 | MUST | P1 | web | TASK-WEB-001 | 8 |
| TASK-WEB-003 | MUST | P1 | web | TASK-WEB-001, TASK-DEAL-003 | 6 |
| TASK-B2B-002 | SHOULD | P3 | b2b | TASK-B2B-001, TASK-BILL-001 | 8 |
| TASK-WEB-005 | SHOULD | P1 | web | TASK-INFRA-001, TASK-WEB-001 | 6 |
| TASK-B2B-003 | COULD | P3 | b2b | TASK-B2B-001 | 8 |
| TASK-B2B-004 | COULD | P3 | b2b | TASK-INFRA-001, TASK-B2B-001 | 6 |
| TASK-MOBILE-003 | COULD | P3 | mobile | TASK-MOBILE-001, TASK-BILL-004 | 6 |

## Layer 5 - 7 task, 49h

| task | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| TASK-CART-004 | MUST | P2 | cart | TASK-CART-003, TASK-INFRA-005 | 6 |
| TASK-DEAL-004 | MUST | P2 | deal | TASK-PRICE-002, TASK-DEAL-002 | 10 |
| TASK-TRACK-002 | MUST | P1 | track | TASK-TRACK-001 | 5 |
| TASK-TRACK-003 | MUST | P1 | track | TASK-TRACK-001 | 6 |
| TASK-TRUST-005 | MUST | P3 | trust | TASK-AFFIL-001, TASK-AFFIL-003, TASK-BILL-002, TASK-TRUST-004 | 6 |
| TASK-MOBILE-002 | SHOULD | P3 | mobile | TASK-MOBILE-001, TASK-CART-003 | 10 |
| TASK-TRUST-006 | SHOULD | P3 | trust | TASK-TRUST-004 | 6 |

## Layer 6 - 6 task, 48h

| task | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| TASK-COMPLY-006 | MUST | P3 | comply | TASK-INFRA-005, TASK-CART-004 | 8 |
| TASK-DEAL-006 | MUST | P2 | deal | TASK-DEAL-004, TASK-TRACK-003 | 6 |
| TASK-TRACK-004 | MUST | P1 | track | TASK-TRACK-003, TASK-PRICE-002, TASK-NOTIF-001 | 6 |
| TASK-WEB-004 | MUST | P1 | web | TASK-WEB-001, TASK-TRACK-002, TASK-TRACK-003 | 6 |
| TASK-AFFIL-005 | SHOULD | P3 | affil | TASK-AFFIL-003, TASK-BILL-002, TASK-TRUST-005 | 10 |
| TASK-DEAL-005 | SHOULD | P2 | deal | TASK-DEAL-004 | 12 |

## Layer 7 - 1 task, 8h

| task | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| TASK-COMPLY-007 | SHOULD | P3 | comply | TASK-COMPLY-001, TASK-COMPLY-006 | 8 |

## Chuỗi tới hạn (critical path) cho MVP

Đường dài nhất tới biểu đồ giá + sale ảo (giá trị lõi MVP):

`TASK-INFRA-002 -> TASK-PRICE-001 -> TASK-PRICE-002 -> TASK-DEAL-001/TASK-DEAL-003 -> TASK-WEB-003`. Song song và cần sớm cho cold-start (tích lũy 90 ngày dữ liệu): `TASK-INFRA-003 -> TASK-SCRAPE-001 -> TASK-SCRAPE-002 -> TASK-SCRAPE-005`. Extension Shopee: `TASK-EXT-001 -> TASK-EXT-002 -> TASK-EXT-003`.

Bắt đầu scraping (TASK-SCRAPE-002) càng sớm càng tốt: tính năng sale ảo (TASK-DEAL-001) cần >=90 ngày lịch sử nên dữ liệu phải chạy nền trước khi UI lên.
