---
fr_id: FR-EXT-006
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-EXT-006 đặc tả UI consent + settings cho extension, bám PDPL §5.5 (đồng thuận tự nguyện/cụ thể/đơn mục đích/tái lập được; im lặng != đồng thuận) và disclosure niềm tin §5.4. Tôi kiểm độc lập: điểm sống còn là consent gate biến đồng thuận trên giấy thành ràng buộc thực thi - mỗi đường dữ liệu (FR-EXT-002 đọc giỏ, FR-EXT-005 đồng bộ) gọi `ensureConsent` trước, chưa opt-in thì chặn. Consent mặc định TẮT (granted rỗng) + granular per-purpose + ConsentRecord tái lập được gửi FR-COMPLY-001. Kết luận: 10/10.

## §2 - Findings

Frontmatter hợp lệ (depends_on [FR-EXT-001, FR-COMPLY-001] đúng; các khóa bắt buộc không rỗng). §1 có 12 mệnh đề; §4 có 12 AC; §5 có 3 test file (consent-default-off/consent-gate/consent-record). Quét typography: glyph `->` duy nhất ở dòng 141 trong code block ts (miễn); prose ASCII thuần, dùng dấu tiếng Việt hợp lệ.

- ISS-001 (kiểm, không phải lỗi): ConsentPurpose ("read_cart"|"read_voucher"|"sync_backend") là kiểu client; consent_record/consent_policy thật thuộc FR-COMPLY-001 trong DATA-MODEL (composite FK purpose_key+policy_version). FR này ghi ConsentRecord rồi `reportConsentToCompliance` - đúng vai bề mặt thu consent, không phát lệnh bảng. Nhất quán DATA-MODEL.
- ISS-002 (kiểm, không phải lỗi): mặc định tắt được AC #2 (getConsent cài mới trả granted: []) + AC #8 (onboarding không checkbox tick sẵn, grep `checked`) khẳng định - đúng "im lặng != đồng thuận" của §5.5.
- ISS-003 (kiểm, không phải lỗi): rút consent hiệu lực ngay (AC #6) + re-consent khi đổi policyVersion (AC #9) phủ quyền chủ thể dữ liệu. Đủ.
- Không phát hiện defect cần sửa trong lượt này.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact (§3/§5) |
|---|---|---|
| #1 consent lúc cài, trước mọi đọc/gửi | #1 | onInstalled -> onboarding + gate |
| #2 granular per-purpose | #3 | ConsentPurpose + granular test |
| #3 im lặng/mặc định-bật != đồng thuận | #2, #8 | consent-default-off + no-checked test |
| #4 gate chặn trước đọc/gửi | #4 | consent-gate.ts ensureConsent + test |
| #5 ConsentRecord tái lập được -> COMPLY-001 | #5 | consent-store.ts + consent-record test |
| #6 disclosure ranh giới kỹ thuật | #7 | onboarding.html nội dung test |
| #7 rút consent hiệu lực ngay | #6 | setConsent([]) + gate test |
| #8 settings xem/đổi + lối DSAR | #11 | settings.ts entry DSAR |
| #9 đổi policy -> re-consent | #9 | policyVersion test |
| #10 consent bền storage, không global | #10 | persist test |
| #11 không dark pattern | #8 | onboarding nút từ chối hiện diện |
| #12 trạng thái trơ trước mọi consent | #1 | gate false khi granted rỗng |

## §4 - Kết luận

Cổng pháp lý PDPL (§5.5) được khóa bằng consent gate + record tái lập được; disclosure (§5.4) kiểm chứng được qua nội dung test. Mọi mệnh đề có AC + test. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập FR-EXT-006.*
