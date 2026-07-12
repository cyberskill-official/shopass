---
fr_id: FR-EXT-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-EXT-004 đặc tả Offscreen API cho DOM parsing/clipboard ngoài service worker, và declarativeNetRequest thay webRequest blocking - hai ràng buộc MV3 nêu rõ trong §3.2 ("dùng declarativeNetRequest thay webRequest blocking; dùng Offscreen API khi cần DOM parsing/clipboard ngoài service worker"). Tôi kiểm độc lập: offscreen vòng đời ngắn (hasDocument trước create, close ngay sau khi xong, trần 10s), DNR static tối thiểu, offscreen KHÔNG tự fetch trang sàn (giữ content script là đường đọc duy nhất). Sau khi sửa một AC thiếu (xem §2), FR đạt 10/10.

## §2 - Findings

Frontmatter hợp lệ (priority SHOULD - hợp lý vì offscreen/DNR chưa bắt buộc cho slice 1 lõi; depends_on FR-EXT-001; các khóa bắt buộc không rỗng). Quét typography: 0 glyph trong prose lẫn code. §5 có 3 test file (offscreen-lifecycle/dnr-rules/offscreen-no-fetch).

- ISS-001 (ĐÃ SỬA): §1 có 12 mệnh đề nhưng §4 ban đầu chỉ có 11 AC, và mệnh đề #9 (message SW <-> offscreen phải typed `ParseDomRequest`/`ParseDomResult` với `target: "offscreen"`) KHÔNG có AC phủ - chỉ AC `tsc sạch` phủ gián tiếp tính đúng kiểu nói chung, không khẳng định hình dạng message + định tuyến `target`. Đã chèn AC mới (nay là #11): "Message SW <-> offscreen là ParseDomRequest/ParseDomResult typed với target: 'offscreen'; message thiếu target hoặc sai kiểu bị tsc bắt", và đẩy AC `npm test` thành #12. File đã sửa: FR-EXT-004-offscreen-declarativenetrequest.md (khối §4 AC).
- ISS-002 (kiểm, không phải lỗi): AC #2 grep no `chrome.webRequest` blocking + AC #6 grep offscreen no `fetch(` tới domain sàn - hai ranh giới cứng của §3.2 đều có test.
- ISS-003 (kiểm, không phải lỗi): không có bảng CSDL phát lệnh; ParseDomResult là kiểu message, không đối chiếu DATA-MODEL.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC (sau sửa) | Artefact (§3/§5) |
|---|---|---|
| #1 DOM/clipboard nặng trong offscreen, reason | #5 | manager.ts createDocument reasons |
| #2 hasDocument trước create (tối đa 1) | #3 | manager.ts + offscreen-lifecycle.test |
| #3 vòng đời ngắn, close ngay | #4 | closeOffscreenDocument + lifecycle test |
| #4 declarativeNetRequest, không webRequest blocking | #2 | grep no webRequest |
| #5 DNR static tối thiểu, mỗi rule có lý do | #8 | rules.json + dnr-rules.test |
| #6 offscreen không tự fetch trang sàn | #6 | offscreen-no-fetch.test |
| #7 kết quả qua minimize() FR-EXT-003 | #7 | đường dẫn parse -> minimize |
| #8 manifest thêm permissions offscreen/DNR | #1 | manifest.json |
| #9 message typed + target offscreen | #11 (mới) | types.ts + tsc |
| #10 createDocument lỗi xử lý lịch sự | #10 | manager tái dùng/báo lỗi |
| #11 dnr.ts sanity check rule | #8 | dnr.ts kiểm số/phạm vi |
| #12 tác vụ offscreen trần thời gian | #9 | sendWithTimeout 10s |

## §4 - Kết luận

Hai ranh giới MV3 (offscreen vòng đời ngắn, DNR thay webRequest blocking) được khóa cứng bằng grep test; lỗ hổng coverage cho mệnh đề #9 đã được vá bằng một AC bổ sung. Score = 10/10. Verdict: PASS.

---

*Hết audit độc lập FR-EXT-004.*
