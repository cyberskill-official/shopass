---
fr_id: FR-MOBILE-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-MOBILE-003 đặc tả deep-link + share-on-sale virality + referral ở mức triển khai được. 11 mệnh đề §1 normative, mỗi mệnh đề có AC và test. Ranh giới chống-spam + chống-gian-lận được giữ chặt: share user-initiated (không tự đăng nền), referral_code lấy từ FR-BILL-004 (không tự sinh), self-referral chặn hai lớp, link công khai chỉ mang product_id + ref ngắn (không token/PII). Trả thưởng + anti-fraud do backend gác cổng. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Virality biến thành spam (đã chốt)
Tự động đăng/gửi link nền lặp đúng kiểu spam affiliate mà §5.7 cấm. Giải: §1 #3 + DEC-MOBILE-22 chỉ tạo link khi bấm Chia sẻ; review + AC #4.

### ISS-002 - Mobile tự sinh mã / tự trả thưởng
Bỏ qua hook anti-fraud (FR-TRUST-004), mở cửa farming. Giải: §1 #2/#8 + DEC-MOBILE-21/24 mã từ FR-BILL-004, thưởng do backend + delay; test `mobile không tự sinh mã`.

### ISS-003 - Rò token/PII qua link công khai
Link đi ra ngoài nên không được mang bí mật. Giải: §1 #7 + DEC-MOBILE-25 chỉ product_id + ref ngắn; test `share link không token/PII`.

### ISS-004 - Self-referral + mất attribution
Tự giới thiệu để farming; attribution mất giữa click và đăng ký. Giải: §1 #5/#6 + DEC-MOBILE-23/24 chặn self-referral hai lớp + lưu ref pending tới đăng ký; test `self-referral bị chặn` + `ref người khác giữ để gắn`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 universal/app link | #1 | `shareLink.ts::buildShareLink` |
| #2 mã từ FR-BILL-004 | #2 | `myReferralCode` + test |
| #3 user-initiated | #4 | `ShareSheet.tsx` + review |
| #4 route + ref pending | #5 | `linkHandler.ts` + test |
| #5 attribution lúc đăng ký | #6 | `authClient.ts` + FR-BILL-004 |
| #6 chặn self-referral | #7 | `pendingReferral.ts` + test |
| #7 không token/PII | #3 | test link không token |
| #8 thưởng do backend | #9 | DEC-MOBILE-24 + review |
| #9 link sai xử lý nhã | #8 | `TestLink thiếu product` |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/test backing. Chống-spam (user-initiated) và chống-gian-lận (mã backend, self-referral hai lớp, thưởng delay) được kiểm bằng test hành vi. Không mệnh đề mồ côi. Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit FR-MOBILE-003.*
