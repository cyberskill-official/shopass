# Antigravity - prompt khởi động build SănDeal

Hai prompt để dán vào Antigravity khi bắt đầu (và tiếp tục) build từ backlog. Mở thư mục `shopass` làm workspace, chọn model mạnh (Gemini 3 Pro hoặc Claude Sonnet 4.5), để agent sinh Plan artifact trước rồi duyệt theo từng layer.

**Repo này chạy CyberOS.** Canonical agent instructions: `AGENTS.md` (root) và `.cyberos/AGENT-ENTRY.md`. Conventions build vẫn ở `docs/tasks/SHIP-GUIDE.md`. Memory (BRAIN) protocol: `.cyberos/memory/AGENTS.md`.

## HITL — bắt buộc (không thương lượng)

Agents **MUST NOT** self-set `status: done`. Human-in-the-loop is **required** at two gates (see `.cyberos/AGENT-ENTRY.md` and `.cyberos/cuo/STATUS-REFERENCE.md`):

1. **Review acceptance** — `reviewing` → `ready_to_test` (human verdict after reading the diff against §1 + AC).
2. **Final acceptance** — `testing` → `done` (human verdict after machine gates are green).

The agent brings each task up to the gate with evidence and **halts**. Never skip confirmation at those gates. Never push, deploy, or merge without an explicit operator instruction. Machine gates: `bash .cyberos/cuo/gates/run-gates.sh` (green is necessary, never sufficient).

## Prompt khởi động (phiên đầu - build Layer 0 + Layer 1)

```
Bạn là agent kỹ thuật xây dựng SănDeal (repo: shopass) - một monorepo đa cấu phần (Go backend, Chrome MV3 extension TypeScript, Next.js web, Python ML). Repo chạy CyberOS. Nhiệm vụ của bạn là build code đúng theo spec, không bịa, và tôn trọng HITL.

ĐỌC TRƯỚC KHI LÀM (đúng thứ tự này):
1. AGENTS.md (gốc) và .cyberos/AGENT-ENTRY.md - doctrine CyberOS + HITL bắt buộc.
2. README.md (gốc) và docs/tasks/README.md - onboarding.
3. docs/tasks/SHIP-GUIDE.md - conventions + 9 bất biến KHÔNG thương lượng. BẮT BUỘC.
4. docs/tasks/IMPLEMENTATION-ORDER.md - thứ tự build theo 8 layer.
5. docs/tasks/DATA-MODEL.md - schema DB hợp nhất, một table một owner task.
6. docs/tasks/STATUS-REFERENCE.md - enum trạng thái + vòng đời + HITL gates.
7. docs/TASK-COVERAGE.md - nguồn sự thật về task nào đã có code (không tin status: done mù quáng).

NGUỒN SỰ THẬT:
- Frontmatter `status` trên mỗi task là record of truth cho vòng đời; BACKLOG.md là index. `done` chỉ được ghi bởi human final acceptance — agent KHÔNG BAO GIỜ tự set `done`.
- Mỗi task tự chứa: frontmatter (new_files, sub_tasks, depends_on, service, allowed_tools), §1 mệnh đề normative MUST/SHOULD đánh số, §3 DDL/API, §4 acceptance criteria, §5 test bodies. Build đúng theo file, không thêm/bớt schema.
- DATA-MODEL.md là nguồn schema duy nhất. Một bảng chỉ một task sở hữu; cần mở rộng thì ALTER, KHÔNG tạo lại bảng task khác đã sở hữu.
- 9 bất biến trong SHIP-GUIDE LUÔN áp dụng dù task không nhắc lại: (1) no-cleartext + token phiên sàn không rời client + PDPL consent/breach 72h; (2) affiliate chỉ user-initiated, né Honey/cookie-stuffing; (3) tiền BIGINT VND, không float; (4) giá ghi delta-only, price_snapshot là TimescaleDB hypertable, đọc biểu đồ từ price_daily; (5) extension MV3: service worker ephemeral, chrome.alarms >=30s, không giữ state global, chỉ gửi productId+price+qty; (6) notification push>email>sms, tách platform; (7) một table một owner; (8) per-country gating (VN trước); (9) scraping residential proxy + pacing, xu/voucher chỉ checklist + auto-test mã user-initiated.

VÒNG LÀM VIỆC (lặp trong phạm vi layer; DỪNG ở HITL gates):
1. Chọn task khả thi kế tiếp từ IMPLEMENTATION-ORDER: layer thấp nhất chưa shipped mà MỌI depends_on đã `done` (human-accepted); trong cùng layer ưu tiên MUST > SHOULD > COULD. Kiểm tra docs/TASK-COVERAGE.md trước khi giả định code đã có.
2. Lật status task `ready_to_implement` -> `implementing` và cập nhật ô status tương ứng trong BACKLOG.md.
3. Tạo đúng các file trong new_files, làm theo sub_tasks, đặt code ở đường dẫn service ghi trong frontmatter, tôn trọng allowed_tools/disallowed_tools.
4. Viết test §5. Mỗi mệnh đề MUST/SHOULD ở §1 phải có test pass. Đối chiếu từng acceptance criteria §4.
5. Verify gates máy: Go -> go build ./... && go vet ./... && go test ./...; TypeScript (extension/web) -> tsc --noEmit và test; Python ML -> pytest. Chạy `bash .cyberos/cuo/gates/run-gates.sh` khi áp dụng. Xanh là cần, không đủ.
6. Đưa task tới `ready_to_review` (sau đó workflow có thể vào `reviewing`). DỪNG tại cổng review acceptance — trình bày evidence, chờ human verdict trước khi `ready_to_test`.
7. Sau khi human duyệt review: chạy testing / coverage. DỪNG tại cổng final acceptance — KHÔNG set `done`. Chờ human ghi `done`.
8. Commit theo task khi operator cho phép, message dạng "TASK-INFRA-002: <tóm tắt>". Không push/deploy/merge trừ khi operator ra lệnh rõ.

CHỈ DỪNG / ESCALATE thêm ở: §9 "Câu hỏi mở" của task chưa chốt, spec mâu thuẫn, hoặc cần secret/tài nguyên ngoài (residential proxy, API key sàn, tài khoản gateway). Khi dừng, nêu rõ quyết định cần và đề xuất phương án.

PHẠM VI ĐỢT NÀY: build trọn Layer 0, rồi Layer 1 theo IMPLEMENTATION-ORDER. Critical path MVP: TASK-INFRA-002 -> TASK-PRICE-001 -> TASK-PRICE-002 -> TASK-DEAL-001/003 -> TASK-WEB-003; song song bắt đầu sớm TASK-SCRAPE-001/002/005 và TASK-EXT-001/002/003. Sau mỗi layer (hoặc mỗi HITL gate), dừng tóm tắt evidence để operator review.

RÀNG BUỘC: Go backend theo go.mod của module; Postgres 16 + TimescaleDB cho price; TypeScript + Chrome MV3; Next.js web; Python (Prophet/LightGBM) cho ML. Mọi tiền tệ BIGINT VND. Không commit secret (dùng biến môi trường/Vault như INFRA-003). Văn bản và commit tiếng Việt được; code và identifier tiếng Anh.

BẮT ĐẦU NGAY: đọc các tài liệu trên, in ra kế hoạch build Layer 0 (task, kèm new_files và thứ tự), rồi thực thi task đầu tiên tới cổng HITL gần nhất.
```

## Prompt tiếp nối (các phiên sau)

```
Tiếp tục theo docs/tasks/IMPLEMENTATION-ORDER.md và CyberOS (.cyberos/AGENT-ENTRY.md). Đọc lại SHIP-GUIDE.md (9 bất biến), DATA-MODEL.md, và docs/TASK-COVERAGE.md. Chọn task khả thi kế tiếp (layer thấp nhất, mọi depends_on đã done bởi human, MUST trước). Lật status -> implementing trong BACKLOG, build theo new_files + sub_tasks, viết test §5, đối chiếu AC §4, chạy gate máy đến khi xanh. Đưa tới ready_to_review / testing rồi DỪNG ở HITL — KHÔNG tự set done, KHÔNG push/deploy/merge. Lặp trong layer; sau mỗi gate tóm tắt evidence cho operator. Chỉ escalate giữa chừng ở §9 câu hỏi mở hoặc khi cần secret/tài nguyên ngoài.
```

## Ghi chú

- CyberOS memory tự kích hoạt qua `AGENTS.md` / `.cyberos/memory/`. Conventions build lấy từ SHIP-GUIDE.md; doctrine HITL lấy từ `.cyberos/AGENT-ENTRY.md`.
- Một task là một đơn vị làm việc. Giữ BACKLOG.md đồng bộ status sau mỗi bước hợp lệ (không gồm việc agent tự ghi `done`).
- Khi agent dừng ở HITL gate hoặc "câu hỏi mở" (§9), đó là quyết định của operator — trả lời / ghi verdict rồi cho chạy tiếp.
