---
fr_id: FR-WEB-005
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file FR-WEB-005 hiện tại. FR đặc tả GraphQL BFF đọc đứng sau gateway. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Bốn ranh giới an toàn được giữ chặt và test: không verify token (gateway đã làm), resolver ủy quyền REST (không chạm DB, giữ kiểm IDOR một chỗ), DataLoader chống N+1, trần độ sâu/độ phức tạp chống DoS. Chỉ Query, ghi giữ một đường REST. BFF tiêu thụ hợp đồng REST TRACK/DEAL: resolver gọi track-svc (wishlist/alert) + deal-svc (chart feed). priority SHOULD nhưng đặc tả đủ chặt để build đúng. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- Không verify token (§1 #1, DEC-WEB-21) - nhận claims từ gateway qua context; có resolver-auth test grep không jwt.verify/jwks.
- Resolver ủy quyền REST track-svc/deal-svc (§1 #2, DEC-WEB-22) - KHÔNG chạm DB; giữ kiểm chủ sở hữu một chỗ (chống IDOR); có AC #2 grep không SQL.
- DataLoader gom một batch (§1 #3, DEC-WEB-23) chống N+1; có dataloader-n1 test đếm đúng một batch.
- Trần max depth + cost (§1 #4, DEC-WEB-24) chống DoS GraphQL; có query-depth test chặn trước resolver.
- Chỉ Query, KHÔNG Mutation ghi (§1 #5, DEC-WEB-25) - ghi giữ một đường REST TRACK-002/003.
- productChart trả hình dạng feed FR-DEAL-003 (§1 #9) - BFF không tự tính verdict/median, chỉ chuyển tiếp.
- Truyền request-id xuống REST (§1 #8) giữ vết xuyên service.
- §10 failure-modes 9 hàng không tầm thường (403/404 REST lộ tài nguyên, mất request-id).
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 sau gateway, không verify | #1 | context.ts + resolver-auth test |
| #2 ủy quyền REST (TRACK/DEAL) | #2 | rest-client.ts + resolvers |
| #3 DataLoader N+1 | #3 | loaders/chart-loader.ts + dataloader-n1 test |
| #4 trần depth/cost | #4 | security/limits.ts + query-depth test |
| #5 chỉ Query | #5 | schema.graphql (không Mutation) |
| #6 schema me/wishlists/chart | #6 | schema.graphql |
| #7 yêu cầu user context | #7 | resolver guard UNAUTHENTICATED |
| #8 truyền request-id | #8 | RestClient |
| #9 chuyển tiếp feed | #9 | resolvers/chart.ts |
| #10 lỗi REST ánh xạ trùng lặp | #10 | error mapping |
| #11 tsc/test | #11 | npm test |
| #12 OTel | - | graphql_rejected_total |

## §4 - Kết luận

Mọi mệnh đề normative có mã/test backing; bốn ranh giới an toàn (token, DB, N+1, DoS) được kiểm chứng bằng test. BFF tiêu thụ hợp đồng REST TRACK/DEAL qua resolver ủy quyền. Là SHOULD nên có thể hoãn sau MVP, nhưng đặc tả đủ chặt để build đúng khi tới lượt. Không mệnh đề mồ côi. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit FR-WEB-005.*
