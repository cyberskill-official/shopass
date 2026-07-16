---
fr_id: TASK-TRUST-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại (không tin .audit cũ). TASK-TRUST-004 đặc tả anti-fraud engine ba lớp tín hiệu (velocity + đồ thị quan hệ recursive CTE + luật ngưỡng cấu hình). Nguyên tắc trung tâm: chấm điểm risk_score 0-100 + gắn cờ reasons[] giải thích được để điều tra, KHÔNG auto-ban; risk cao chỉ HOLD reward (không tịch thu). 12 mệnh đề §1 (11 MUST + 1 MUST NOT auto-ban). Schema khớp DATA-MODEL.md (fraud_signal subject_user_id/kind/risk_score SMALLINT CHECK 0..100/reasons JSONB/status CHECK + UNIQUE(subject_user_id,kind); account_link_edge a_user/b_user/link_type/weight REAL PK + CHECK a_user<b_user). Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; phase P3 hợp lệ; key đủ; depends_on=[TASK-BILL-004, TASK-AFFIL-001], blocks=[TASK-TRUST-005, TASK-TRUST-006]. Pass.
- ISS-002 (contract D): fraud_signal + account_link_edge DDL khớp DATA-MODEL exactly; risk_score SMALLINT CHECK BETWEEN 0 AND 100; status CHECK open|investigating|confirmed_fraud|cleared; UNIQUE idempotent; CHECK(a_user<b_user) chuẩn hóa cạnh vô hướng. Không PII sàn (FK nội bộ, §1 #11). Khớp source §5.3/§9. Pass.
- ISS-003 (normative B): clause #2 MUST NOT auto-ban/seize, #6 HOLD reward không tịch thu, #8 status do điều tra không engine, #10 công bằng user lẻ - tiêu chí rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestVelocity_BurstSignup_Flags/NormalPace_Clean, TestGraph_DenseCluster_Flags/SoloUser_NoFlag, TestScore_HighRisk_HoldsNotSeizes (HOLD + KHÔNG banned). Pass.
- ISS-005 (typography O): SQL/Go comment §3 (code block); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 risk_score + reasons | #1,#9 | score.go + fraud_signal.reasons |
| #2 không auto-ban | #2 | TestScore_HighRisk_HoldsNotSeizes |
| #3 velocity | #3,#4 | velocity.go + TestVelocity_Burst/NormalPace |
| #4 đồ thị quan hệ | #5,#6 | graph.go recursive CTE + TestGraph_DenseCluster/SoloUser |
| #5 luật ngưỡng cấu hình | #7 | rules.go + HoldThreshold config |
| #6 HOLD reward | #8 | holdRewards |
| #7 tiêu thụ device TRUST-006 | - | account_link_edge link_type device |
| #8 vòng đời status | #1 | fraud_signal.status CHECK |
| #9 idempotent | #10 | UNIQUE(subject,kind) + upsert |
| #10 công bằng cơ bản | #11 | TestVelocity_NormalPace_Clean |
| #11 không PII | #12 | FK nội bộ |
| #12 metric | #12 | fraud_signal_open_total |

## §4 - Kết luận

Nguyên tắc không-auto-ban giữ nhất quán + test; ba lớp tín hiệu bổ sung; reasons[] giải thích được; schema khớp DATA-MODEL exactly. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-TRUST-004.*
