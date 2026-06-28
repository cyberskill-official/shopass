---
fr_id: FR-EXT-001
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-EXT-001 đặc tả scaffold Manifest V3 và vòng đời service worker ephemeral ở mức triển khai được. Tôi đối chiếu lại từng mệnh đề §1 với §3.2 tài liệu nguồn: ba ràng buộc nền (SW ephemeral nên state ở chrome.storage; chrome.alarms tối thiểu 0.5 phút thay setInterval; host_permissions per-domain thay <all_urls>) đều được khai báo normative và có AC + test backing. Hợp đồng TypeScript cụ thể (storage.ts là điểm state duy nhất, listener top-level đồng bộ, minimum_chrome_version 120). Kết luận độc lập: đạt 10/10.

## §2 - Findings

Tôi kiểm lại frontmatter (id khớp tên file, module EXT, priority MUST, status ready_to_implement, new_files/sub_tasks/risk_if_skipped đều không rỗng) - hợp lệ. Quét typography toàn file: mọi ký tự mũi tên (->) chỉ nằm trong code block jsonc/ts (được miễn); prose dùng ASCII thuần, không em-dash/en-dash/curly quote. Đã đếm: §1 có 12 mệnh đề, §4 có 12 AC, §5 có 3 test file (lifecycle/alarms/manifest), §10 có 9 dòng failure-mode không tầm thường.

- ISS-001 (kiểm, không phải lỗi): §1 #5 ràng `periodInMinutes >= 0.5` đúng mốc Chrome 120 trong §3.2 ("tối thiểu 30 giây"). AC #4 + alarms.test khẳng định. Không sửa.
- ISS-002 (kiểm, không phải lỗi): không có schema CSDL ở FR này nên không có gì để đối chiếu DATA-MODEL; tiền tệ không xuất hiện. Đúng phạm vi scaffold.
- Không phát hiện defect cần sửa trong lượt này. File sạch.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact (§3/§5) |
|---|---|---|
| #1 MV3 + service_worker, không background.page | #1 | manifest.json + manifest.test.ts |
| #2 SW ephemeral, không state global | #5, #6 | storage.ts + lifecycle.test.ts |
| #3 storage.ts điểm state duy nhất | #5 | storage.ts getState/setState |
| #4 rehydrate khi wake | #5 | alarms.ts onAlarm (getState trước) |
| #5 alarms >=0.5 phút, không setInterval | #4, #7 | alarms.ts + alarms.test.ts |
| #6 host_permissions per-domain | #2 | manifest.json host_permissions |
| #7 permissions tối thiểu | #3 | manifest.test.ts |
| #8 việc nặng off-device, không >5 phút | (ràng buộc) | disallowed_tools + §10 |
| #9 messaging.ts typed | #12 | messaging.ts + tsc |
| #10 build tái lập | #9 | build.mjs (esbuild xác định) |
| #11 listener top-level đồng bộ | #8 | service-worker.ts entrypoint |
| #12 minimum_chrome_version 120 | #10 | manifest.json |

## §4 - Kết luận

Mọi mệnh đề normative có manifest/code/test backing; không mệnh đề mồ côi. Ba ràng buộc MV3 truy vết tới NFR-EXT-001. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập FR-EXT-001.*
