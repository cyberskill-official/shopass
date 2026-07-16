---
fr_id: TASK-INFRA-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập, re-derive từ nội dung hiện tại của file (không tin .audit cũ). TASK-INFRA-001 đặc tả API Gateway/BFF ở mức triển khai được: 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD cho OTel metric), mỗi mệnh đề testable, đều có AC §4 và test §5 backing. Tách phát-hành (AUTH) vs verify (gateway) đúng ranh giới zero-trust; rate-limit hai trục + WAF edge + request-id propagation đầy đủ. Code Go thật (router chain, jwtVerify, rateLimit). Đạt 10/10, PASS.

## §2 - Findings

Kiểm độc lập theo từng dimension, không phát hiện defect cần sửa surgical.

- ISS-001 (frontmatter A): kiểm `id`=TASK-INFRA-001 khớp tên file; `module`=INFRA khớp folder; mọi key bắt buộc có và không rỗng (priority MUST, status ready_to_implement, phase P0, slice 1, owner, created, depends_on=[], blocks, effort_hours 8, source_pages, source_decisions, language, service, new_files, sub_tasks, risk_if_skipped). Pass.
- ISS-002 (normative B): 12 clause dùng MUST/SHOULD đúng, đều có tiêu chí kiểm (status code 401/429/413/400, header Retry-After, UUIDv4). Không có "xu ly muot" mơ hồ. Có >=1 MUST. Pass.
- ISS-003 (contract D): code Go thật, middleware chain thứ tự requestID->waf->rateLimit->jwtVerify đúng; verify exp/iss/aud + refresh JWKS theo kid lạ; không tự ký token (DEC-INFRA-02). Không chạm DATA-MODEL (task hạ tầng, không bảng). Pass.
- ISS-004 (AC/test E,F): 14 AC ánh xạ rõ; test JWT expired/bad-aud/valid-propagate, rate-limit per-IP/per-user-isolated, WAF path traversal, request-id generated. Mỗi clause testable có test tên. Pass.
- ISS-005 (typography O): prose ASCII thuần; mũi tên unicode chỉ trong comment Go (§3) - thuộc code block, ngoài phạm vi rule O. Không banned word. Pass.
- ISS-006 (§6-§11 + sentinel M, N): đủ khung triển khai, phụ thuộc khớp frontmatter, payload, open questions, failure-modes (9 dòng non-trivial), notes; sentinel `*Hết TASK-INFRA-001...*` có; new_files/sub_tasks/risk non-empty, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact (dựng từ file hiện tại)

| §1 clause | AC | Artefact §5/§3 |
|---|---|---|
| #1 định tuyến prefix | #1 | router.go NewHandler mux |
| #2 verify JWT JWKS | #6 | jwt.go jwtVerify + TestJWT_Valid_PropagatesUserID |
| #3 từ chối 401 | #2,#3,#4,#5 | TestJWT_Expired/BadAudience_401 |
| #4 inject X-User-Id | #6 | jwtVerify Header.Set + test propagate |
| #5 rate-limit 2 trục | #7,#8 | ratelimit.go + TestRateLimit_PerIP/PerUser |
| #6 ngưỡng per-route-class | #7 | bucketKeyAndLimit |
| #7 X-Request-Id propagate | #12 | requestid.go + TestRequestID_Generated |
| #8 WAF edge | #9,#10,#11 | waf.go + TestWAF_PathTraversal_400 |
| #9 WSS handshake verify | #13 | wsUpgrade |
| #10 secret qua INFRA-003 | - | §6 + disallowed_tools |
| #11 OTel metric (SHOULD) | #3 | metrics.JWTRejected/RateLimited |
| #12 cache JWKS refresh kid | #14 | jwks.Verify refresh |

## §4 - Kết luận

Mọi mệnh đề normative có AC + code/test backing; không mệnh đề mồ côi; phụ thuộc nhất quán với frontmatter; PDPL không áp dụng trực tiếp (task hạ tầng). Score = 10/10. Verdict: PASS. Sẵn sàng build (ready_to_implement).

---

*Hết audit TASK-INFRA-001.*
