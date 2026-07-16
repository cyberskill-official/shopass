---
fr_id: TASK-INFRA-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại. TASK-INFRA-002 đặt nền migration (golang-migrate up/down bất biến), bảng `platform`, cột lõi `app_user`, tài liệu quy ước. 12 mệnh đề §1 (11 MUST + 1 SHOULD log), testable. DDL khớp DATA-MODEL.md + source §3.4: platform (id SMALLINT PK, code/country CHECK), app_user (id BIGSERIAL, email CITEXT UNIQUE, locale 'vi-VN', status). Ranh giới cột lõi (INFRA) vs cột bảo mật (AUTH-001 thêm pwd_hash) rõ. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; mọi key bắt buộc đủ và non-empty; depends_on=[] (nền móng), blocks liệt kê 7 task downstream. Pass.
- ISS-002 (contract D): DDL platform khớp DATA-MODEL owner (TASK-INFRA-002), CHECK code IN (shopee|tiktok|lazada) + country `^[A-Z]{2}$` ISO alpha-2; app_user email CITEXT UNIQUE đúng §3.4. Seed idempotent ON CONFLICT(code). Khoá chính BIGSERIAL cho bảng lớn, SMALLINT cho platform - khớp §1 #11. Pass.
- ISS-003 (normative B): clause #2 thứ tự extension trước, #4 CHECK constraint, #7 CITEXT case-insensitive, #9 runner Up/Down/Version idempotent - đều có tiêu chí kiểm. Pass.
- ISS-004 (AC/test E,F): 12 AC; test TestUp_FromZero/Idempotent, TestDown_OneStep, TestPlatform_CodeCheck/CountryCheck, TestAppUser_EmailCaseInsensitiveUnique/Defaults. Mỗi clause testable có test. Pass.
- ISS-005 (typography O): prose ASCII thuần, không unicode trong prose, không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 9 dòng; sentinel có; new_files/sub_tasks/risk non-empty, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 golang-migrate up/down bất biến | #1,#2,#3 | migrate.go + TestUp/Down |
| #2 0001_extensions trước | #11 | 0001_extensions.up.sql |
| #3 bảng platform §3.4 | #1 | 0002_platform.up.sql |
| #4 CHECK code/country | #6,#7 | CHECK + TestPlatform_Code/CountryCheck |
| #5 seed 3 sàn idempotent | #4,#5 | 0001_platform_seed.sql + TestPlatformSeed_Idempotent |
| #6 app_user cột lõi | #1,#9 | 0003_app_user_core.up.sql + TestAppUser_Defaults |
| #7 email CITEXT unique | #8 | TestAppUser_EmailCaseInsensitiveUnique |
| #8 AUTH-001 mở rộng sau | - | comment migration + §7 |
| #9 runner Up/Down/Version | #2,#3,#10 | migrate.go Up/Down/Version |
| #10 quy ước đặt tên (doc) | #12 | NAMING-AND-MIGRATIONS.md |
| #11 BIGSERIAL bảng lớn | - | DDL types |
| #12 log migration (SHOULD) | - | §1 #12 |

## §4 - Kết luận

DDL khớp catalog + source §3.4; CHECK/CITEXT kiểm bằng test thật (không mock); ranh giới module sạch. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-INFRA-002.*
