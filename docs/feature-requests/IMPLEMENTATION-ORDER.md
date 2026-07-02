# SănDeal - Thứ tự triển khai (build order)

Tài liệu này dẫn xuất tự động từ đồ thị `depends_on` của 90 FR. Một agent triển khai nên build theo **layer**: mọi FR trong layer N chỉ phụ thuộc vào FR ở layer < N, nên trong cùng một layer các FR build song song được. DAG đã được xác minh **acyclic** (không có phụ thuộc vòng) và **reciprocal** (`blocks` là nghịch đảo của `depends_on`).

Tổng: **90 FR**, **627h**, **8 layer**. Mỗi FR ở status `done` (audit 10/10).

## Cách dùng cho agent ship

1. Lấy FR khả thi tiếp theo = FR ở layer thấp nhất chưa `done` mà mọi `depends_on` đã `done`.
2. Đọc file FR (frontmatter + §1-§11), build theo `new_files`/`sub_tasks`, chạy test §5, đối chiếu acceptance criteria §4.
3. Cập nhật status theo `STATUS-REFERENCE.md` (`implementing` -> `ready_to_review` -> ... -> `done`).
4. Ưu tiên trong cùng layer: `MUST` trước `SHOULD` trước `COULD`.

## Layer 0 - 4 FR, 27h

| FR | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| FR-EXT-001 | MUST | P1 | ext | - | 8 |
| FR-INFRA-001 | MUST | P0 | infra | - | 8 |
| FR-INFRA-002 | MUST | P0 | infra | - | 6 |
| FR-INFRA-003 | MUST | P0 | infra | - | 5 |

## Layer 1 - 12 FR, 78h

| FR | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| FR-AFFIL-001 | MUST | P2 | affil | FR-INFRA-002 | 6 |
| FR-AUTH-001 | MUST | P1 | auth | FR-INFRA-002 | 6 |
| FR-CART-001 | MUST | P2 | cart | FR-INFRA-002 | 6 |
| FR-COMPLY-001 | MUST | P1 | comply | FR-INFRA-002 | 8 |
| FR-COMPLY-005 | MUST | P1 | comply | FR-INFRA-003 | 5 |
| FR-EXT-002 | MUST | P1 | ext | FR-EXT-001 | 10 |
| FR-INFRA-004 | MUST | P0 | infra | FR-INFRA-001 | 8 |
| FR-INFRA-005 | MUST | P0 | infra | FR-INFRA-002 | 6 |
| FR-NOTIF-001 | MUST | P1 | notif | FR-INFRA-002 | 6 |
| FR-PRICE-001 | MUST | P1 | price | FR-INFRA-002 | 6 |
| FR-TRUST-001 | MUST | P1 | trust | FR-EXT-001 | 6 |
| FR-EXT-004 | SHOULD | P1 | ext | FR-EXT-001 | 5 |

## Layer 2 - 19 FR, 130h

| FR | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| FR-AFFIL-002 | MUST | P2 | affil | FR-AFFIL-001 | 6 |
| FR-AUTH-002 | MUST | P1 | auth | FR-AUTH-001, FR-INFRA-001 | 6 |
| FR-AUTH-003 | MUST | P1 | auth | FR-AUTH-001 | 5 |
| FR-BILL-001 | MUST | P2 | bill | FR-AUTH-001 | 6 |
| FR-CART-005 | MUST | P2 | cart | FR-EXT-002, FR-CART-001 | 6 |
| FR-COMPLY-002 | MUST | P1 | comply | FR-COMPLY-001 | 6 |
| FR-COMPLY-003 | MUST | P1 | comply | FR-COMPLY-001 | 8 |
| FR-COMPLY-004 | MUST | P1 | comply | FR-COMPLY-001, FR-INFRA-004 | 5 |
| FR-EXT-003 | MUST | P1 | ext | FR-EXT-002 | 6 |
| FR-EXT-006 | MUST | P1 | ext | FR-EXT-001, FR-COMPLY-001 | 5 |
| FR-EXT-007 | MUST | P2 | ext | FR-EXT-002 | 10 |
| FR-EXT-008 | MUST | P2 | ext | FR-EXT-002 | 8 |
| FR-NOTIF-002 | MUST | P1 | notif | FR-NOTIF-001 | 8 |
| FR-NOTIF-003 | MUST | P1 | notif | FR-NOTIF-001 | 8 |
| FR-PRICE-002 | MUST | P1 | price | FR-PRICE-001 | 8 |
| FR-PRICE-005 | MUST | P1 | price | FR-PRICE-001 | 8 |
| FR-SCRAPE-001 | MUST | P1 | scrape | FR-INFRA-003, FR-PRICE-001 | 10 |
| FR-CART-006 | SHOULD | P2 | cart | FR-EXT-002 | 5 |
| FR-COMPLY-008 | SHOULD | P3 | comply | FR-COMPLY-001 | 6 |

## Layer 3 - 24 FR, 162h

| FR | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| FR-AFFIL-003 | MUST | P2 | affil | FR-AFFIL-002 | 8 |
| FR-AFFIL-004 | MUST | P2 | affil | FR-AFFIL-002, FR-EXT-003 | 5 |
| FR-AUTH-005 | MUST | P1 | auth | FR-AUTH-001, FR-COMPLY-003 | 5 |
| FR-BILL-002 | MUST | P2 | bill | FR-BILL-001 | 10 |
| FR-CART-002 | MUST | P2 | cart | FR-EXT-003 | 5 |
| FR-DEAL-001 | MUST | P1 | deal | FR-PRICE-002 | 8 |
| FR-DEAL-003 | MUST | P1 | deal | FR-PRICE-002 | 5 |
| FR-EXT-005 | MUST | P1 | ext | FR-EXT-003, FR-AUTH-002 | 6 |
| FR-NOTIF-004 | MUST | P1 | notif | FR-NOTIF-003 | 6 |
| FR-NOTIF-005 | MUST | P2 | notif | FR-NOTIF-003 | 5 |
| FR-NOTIF-006 | MUST | P2 | notif | FR-NOTIF-003 | 4 |
| FR-PRICE-003 | MUST | P1 | price | FR-PRICE-002 | 5 |
| FR-PRICE-004 | MUST | P1 | price | FR-PRICE-005 | 6 |
| FR-SCRAPE-002 | MUST | P1 | scrape | FR-SCRAPE-001 | 8 |
| FR-SCRAPE-003 | MUST | P1 | scrape | FR-SCRAPE-001 | 12 |
| FR-TRUST-002 | MUST | P1 | trust | FR-EXT-003 | 5 |
| FR-TRUST-003 | MUST | P1 | trust | FR-EXT-003, FR-COMPLY-005 | 6 |
| FR-WEB-001 | MUST | P1 | web | FR-AUTH-002 | 8 |
| FR-AUTH-004 | SHOULD | P1 | auth | FR-AUTH-002 | 6 |
| FR-B2B-001 | SHOULD | P3 | b2b | FR-PRICE-002, FR-COMPLY-003 | 10 |
| FR-BILL-004 | SHOULD | P2 | bill | FR-BILL-001 | 5 |
| FR-BILL-005 | SHOULD | P2 | bill | FR-BILL-001 | 6 |
| FR-MOBILE-001 | SHOULD | P3 | mobile | FR-AUTH-002, FR-NOTIF-002 | 12 |
| FR-NOTIF-007 | SHOULD | P2 | notif | FR-NOTIF-003 | 6 |

## Layer 4 - 17 FR, 125h

| FR | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| FR-BILL-003 | MUST | P2 | bill | FR-BILL-002 | 6 |
| FR-CART-003 | MUST | P2 | cart | FR-CART-001, FR-CART-002 | 12 |
| FR-DEAL-002 | MUST | P1 | deal | FR-DEAL-001 | 6 |
| FR-SCRAPE-004 | MUST | P1 | scrape | FR-SCRAPE-003 | 8 |
| FR-SCRAPE-005 | MUST | P1 | scrape | FR-SCRAPE-002, FR-PRICE-002 | 6 |
| FR-SCRAPE-006 | MUST | P1 | scrape | FR-SCRAPE-002 | 6 |
| FR-SCRAPE-007 | MUST | P2 | scrape | FR-SCRAPE-003 | 10 |
| FR-SCRAPE-008 | MUST | P2 | scrape | FR-SCRAPE-003 | 8 |
| FR-TRACK-001 | MUST | P1 | track | FR-PRICE-001, FR-SCRAPE-002 | 5 |
| FR-TRUST-004 | MUST | P3 | trust | FR-BILL-004, FR-AFFIL-001 | 10 |
| FR-WEB-002 | MUST | P1 | web | FR-WEB-001 | 8 |
| FR-WEB-003 | MUST | P1 | web | FR-WEB-001, FR-DEAL-003 | 6 |
| FR-B2B-002 | SHOULD | P3 | b2b | FR-B2B-001, FR-BILL-001 | 8 |
| FR-WEB-005 | SHOULD | P1 | web | FR-INFRA-001, FR-WEB-001 | 6 |
| FR-B2B-003 | COULD | P3 | b2b | FR-B2B-001 | 8 |
| FR-B2B-004 | COULD | P3 | b2b | FR-INFRA-001, FR-B2B-001 | 6 |
| FR-MOBILE-003 | COULD | P3 | mobile | FR-MOBILE-001, FR-BILL-004 | 6 |

## Layer 5 - 7 FR, 49h

| FR | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| FR-CART-004 | MUST | P2 | cart | FR-CART-003, FR-INFRA-005 | 6 |
| FR-DEAL-004 | MUST | P2 | deal | FR-PRICE-002, FR-DEAL-002 | 10 |
| FR-TRACK-002 | MUST | P1 | track | FR-TRACK-001 | 5 |
| FR-TRACK-003 | MUST | P1 | track | FR-TRACK-001 | 6 |
| FR-TRUST-005 | MUST | P3 | trust | FR-AFFIL-001, FR-AFFIL-003, FR-BILL-002, FR-TRUST-004 | 6 |
| FR-MOBILE-002 | SHOULD | P3 | mobile | FR-MOBILE-001, FR-CART-003 | 10 |
| FR-TRUST-006 | SHOULD | P3 | trust | FR-TRUST-004 | 6 |

## Layer 6 - 6 FR, 48h

| FR | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| FR-COMPLY-006 | MUST | P3 | comply | FR-INFRA-005, FR-CART-004 | 8 |
| FR-DEAL-006 | MUST | P2 | deal | FR-DEAL-004, FR-TRACK-003 | 6 |
| FR-TRACK-004 | MUST | P1 | track | FR-TRACK-003, FR-PRICE-002, FR-NOTIF-001 | 6 |
| FR-WEB-004 | MUST | P1 | web | FR-WEB-001, FR-TRACK-002, FR-TRACK-003 | 6 |
| FR-AFFIL-005 | SHOULD | P3 | affil | FR-AFFIL-003, FR-BILL-002, FR-TRUST-005 | 10 |
| FR-DEAL-005 | SHOULD | P2 | deal | FR-DEAL-004 | 12 |

## Layer 7 - 1 FR, 8h

| FR | Pri | Phase | Module | depends_on | h |
|---|:-:|:-:|---|---|--:|
| FR-COMPLY-007 | SHOULD | P3 | comply | FR-COMPLY-001, FR-COMPLY-006 | 8 |

## Chuỗi tới hạn (critical path) cho MVP

Đường dài nhất tới biểu đồ giá + sale ảo (giá trị lõi MVP):

`FR-INFRA-002 -> FR-PRICE-001 -> FR-PRICE-002 -> FR-DEAL-001/FR-DEAL-003 -> FR-WEB-003`. Song song và cần sớm cho cold-start (tích lũy 90 ngày dữ liệu): `FR-INFRA-003 -> FR-SCRAPE-001 -> FR-SCRAPE-002 -> FR-SCRAPE-005`. Extension Shopee: `FR-EXT-001 -> FR-EXT-002 -> FR-EXT-003`.

Bắt đầu scraping (FR-SCRAPE-002) càng sớm càng tốt: tính năng sale ảo (FR-DEAL-001) cần >=90 ngày lịch sử nên dữ liệu phải chạy nền trước khi UI lên.
