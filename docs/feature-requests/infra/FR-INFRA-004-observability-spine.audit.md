---
fr_id: FR-INFRA-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. FR-INFRA-004 đặc tả xương sống quan sát: OTel tracing xuyên service (W3C traceparent, trace_id bắt nguồn từ X-Request-Id của gateway), Prometheus pull /metrics bảo vệ, log JSON bắt buộc trace_id/request_id, mask PII theo PDPL. 12 mệnh đề §1 (10 MUST + 1 SHOULD sampling + 1 MUST dashboard), testable; có test trace nối liền hai service và test redact PII. depends_on=[FR-INFRA-001] hợp lý (cần X-Request-Id). Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; key bắt buộc đủ; depends_on=[FR-INFRA-001], blocks=[FR-COMPLY-004] nhất quán §7. Pass.
- ISS-002 (contract D): code Go thật - InitTracer + propagation.TraceContext (W3C), middleware Extract->Start->inject, redactAttrs mask email/phone/token/cookie/authorization. Khớp source §3.8 (Prometheus+Grafana+OTel+structured logs). Pass.
- ISS-003 (normative + PDPL B,P): clause #6 mask PII (DEC-INFRA-20, đúng tinh thần PDPL: log không chứa dữ liệu cá nhân thô), #2 không tạo trace_id mới giữa chừng, #12 dashboard vẽ ngưỡng NFR. Tiêu chí kiểm rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestTrace_ContinuesAcrossServices (cùng trace_id), TestLog_HasTraceAndRequestID, TestLog_RedactsPII, TestMetrics_HTTPObserve. Mỗi clause có test. Pass.
- ISS-005 (typography O): prose ASCII thuần, không unicode prose, không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 9 dòng (gồm clock skew, high-cardinality); sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 OTel tracing | #2 | tracing.go InitTracer |
| #2 traceparent nối liền | #2,#3 | TestTrace_ContinuesAcrossServices |
| #3 /metrics bảo vệ | #1,#9 | metrics.go + handler |
| #4 helper counter/histogram | #7 | HTTPObserve + TestMetrics_HTTPObserve |
| #5 log JSON trace_id | #4 | logging.go + TestLog_HasTraceAndRequestID |
| #6 mask PII | #5,#6 | redact.go + TestLog_RedactsPII |
| #7 middleware đầu chain | #2 | middleware.go HTTP |
| #8 inject traceparent | #8 | middleware Inject |
| #9 Grafana dashboard | #11 | overview.json |
| #10 sampling (SHOULD) | - | errorBiasedSampler |
| #11 shutdown flush | #10 | tp.Shutdown |
| #12 dashboard ngưỡng NFR | #12 | overview.json |

## §4 - Kết luận

Mọi mệnh đề có code/test backing gồm trace cross-service và redact PII; mask PII đúng tinh thần PDPL; nối FR-COMPLY-004 breach. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit FR-INFRA-004.*
