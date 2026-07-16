---
fr_id: TASK-TRUST-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Audit độc lập từ nội dung hiện tại (không tin .audit cũ). TASK-TRUST-001 đặc tả extension mã nguồn mở + reproducible build + disclosure. Trục: reproducible build so SHA-256 bản ship với build-lại-từ-source (build-deterministic.mjs + verify-reproducible.sh, SOURCE_DATE_EPOCH + zip xác định). Disclosure neo vào allowlist TASK-EXT-003 qua disclosure-consistency.test.ts (không lệch hành vi thật). CI publish gate biến ba bất biến (tái lập + có tag + disclosure khớp) thành điều kiện ship. 12 mệnh đề §1 (10 MUST + 1 MUST NOT obfuscation + 1 SHOULD hash). Không bảng DB (task doc/build/test). Nền cho TRUST-003. Đạt 10/10, PASS.

## §2 - Findings

- ISS-001 (frontmatter A): id/module/folder khớp; key đủ; depends_on=[TASK-EXT-001]. related_frs có NFR-TRUST-001 (hợp lệ, related có thể trỏ NFR). Pass.
- ISS-002 (contract D): script thật - verify-reproducible.sh build lại trong worktree sạch + so shasum, exit 1 nếu lệch; build-deterministic.mjs loại timestamp/path; DISCLOSURE.md liệt kê đúng tập {platform, productId, price, qty} + KHÔNG cookie/mật khẩu/token. Khớp source §5.4. Pass.
- ISS-003 (normative B): clause #2 deterministic, #6 disclosure đúng tập, #8 test consistency, #11 MUST NOT obfuscation - tiêu chí kiểm rõ. Pass.
- ISS-004 (AC/test E,F): 12 AC; test disclosure-consistency.test.ts (trường allowlist phải khai; trường mới chưa khai -> fail) + bash verify-reproducible + diff hai lần build. Pass.
- ISS-005 (typography O): dấu tiếng Việt trong DISCLOSURE.md template + comment §3 (code block/markdown nhúng, scoped out rule O); prose ASCII thuần; không banned word. Pass.
- ISS-006 (§6-§11, M, N): đủ khung; failure-modes 8 dòng; sentinel có; self-contained, không TBD. Pass.

## §3 - Traceability §1 -> AC -> artefact

| §1 clause | AC | Artefact |
|---|---|---|
| #1 source công khai + tag | #1,#2 | repo + manifest homepage_url/version |
| #2 reproducible build | #3,#10 | build-deterministic.mjs |
| #3 pin toolchain | #10 | lockfile + REPRODUCIBLE-BUILD.md |
| #4 verify hash | #4,#5 | verify-reproducible.sh |
| #5 CI publish gate | #9 | reproducible-publish-gate.yml |
| #6 disclosure đúng tập | #6 | DISCLOSURE.md |
| #7 listing nhất quán | #8 | chrome-web-store-listing.md |
| #8 disclosure parity test | #7 | disclosure-consistency.test.ts |
| #9 manifest homepage/version | #2 | manifest.json |
| #10 disclosure nêu lý do quyền | #6 | DISCLOSURE.md §quyền |
| #11 cấm obfuscation | #11 | review + grep bundle |
| #12 publish hash (SHOULD) | #12 | release notes |

## §4 - Kết luận

Reproducible build + disclosure-parity-test biến lời hứa thành kiểm chứng được; CI gate giữ ba bất biến; nền cho audit độc lập TRUST-003. Không defect. Score = 10/10. Verdict: PASS.

---

*Hết audit TASK-TRUST-001.*
