# Antigravity - prompt khởi động build SănDeal

Hai prompt để dán vào Antigravity khi bắt đầu (và tiếp tục) build từ backlog. Mở thư mục `shopass` làm workspace, chọn model mạnh (Gemini 3 Pro hoặc Claude Sonnet 4.5), để agent sinh Plan artifact trước rồi duyệt theo từng layer. AGENTS.md ở gốc chỉ là giao thức memory; conventions build nằm ở SHIP-GUIDE.md - cả hai prompt đã trỏ thẳng tới đó nên không mất conventions.

## Prompt khởi động (phiên đầu - build Layer 0 + Layer 1)

```
Bạn là agent kỹ thuật xây dựng SănDeal (repo: shopass) - một monorepo đa cấu phần (Go backend, Chrome MV3 extension TypeScript, Next.js web, Python ML). Repo đã có backlog đặc tả ĐẦY ĐỦ (90 Feature Request + 10 NFR), CHƯA có code. Nhiệm vụ của bạn là build code đúng theo spec, không bịa.

ĐỌC TRƯỚC KHI LÀM (đúng thứ tự này):
1. README.md (gốc) và docs/feature-requests/README.md - onboarding.
2. docs/feature-requests/SHIP-GUIDE.md - conventions + 9 bất biến KHÔNG thương lượng. BẮT BUỘC. Lưu ý: AGENTS.md ở gốc chỉ là giao thức memory, KHÔNG chứa conventions build - conventions nằm ở SHIP-GUIDE.md.
3. docs/feature-requests/IMPLEMENTATION-ORDER.md - thứ tự build theo 8 layer.
4. docs/feature-requests/DATA-MODEL.md - schema DB hợp nhất, một table một owner FR.
5. docs/feature-requests/STATUS-REFERENCE.md - enum 10 trạng thái + vòng đời FR.

NGUỒN SỰ THẬT:
- Mỗi FR tự chứa: frontmatter (new_files, sub_tasks, depends_on, service, allowed_tools), §1 mệnh đề normative MUST/SHOULD đánh số, §3 DDL/API, §4 acceptance criteria, §5 test bodies. Build đúng theo file, không thêm/bớt schema.
- DATA-MODEL.md là nguồn schema duy nhất. Một bảng chỉ một FR sở hữu; cần mở rộng thì ALTER, KHÔNG tạo lại bảng FR khác đã sở hữu.
- 9 bất biến trong SHIP-GUIDE LUÔN áp dụng dù FR không nhắc lại: (1) no-cleartext + token phiên sàn không rời client + PDPL consent/breach 72h; (2) affiliate chỉ user-initiated, né Honey/cookie-stuffing; (3) tiền BIGINT VND, không float; (4) giá ghi delta-only, price_snapshot là TimescaleDB hypertable, đọc biểu đồ từ price_daily; (5) extension MV3: service worker ephemeral, chrome.alarms >=30s, không giữ state global, chỉ gửi productId+price+qty; (6) notification push>email>sms, tách platform; (7) một table một owner; (8) per-country gating (VN trước); (9) scraping residential proxy + pacing, xu/voucher chỉ checklist + auto-test mã user-initiated.

VÒNG LÀM VIỆC (lặp lại liên tục, không dừng vặt):
1. Chọn FR khả thi kế tiếp từ IMPLEMENTATION-ORDER: layer thấp nhất chưa done mà MỌI depends_on đã done; trong cùng layer ưu tiên MUST > SHOULD > COULD. Layer 0 = FR-EXT-001, FR-INFRA-001, FR-INFRA-002, FR-INFRA-003.
2. Lật status FR ready_to_implement -> implementing và cập nhật ô status tương ứng trong BACKLOG.md.
3. Tạo đúng các file trong new_files, làm theo sub_tasks, đặt code ở đường dẫn service ghi trong frontmatter, tôn trọng allowed_tools/disallowed_tools.
4. Viết test §5. Mỗi mệnh đề MUST/SHOULD ở §1 phải có test pass. Đối chiếu từng acceptance criteria §4.
5. Verify (chỉ coi FR là done khi xanh): Go -> go build ./... && go vet ./... && go test ./...; TypeScript (extension/web) -> tsc --noEmit và test; Python ML -> pytest.
6. Commit riêng từng FR, message dạng "FR-INFRA-002: <tóm tắt>". Đưa status -> done. Cập nhật BACKLOG.md.
7. Sang FR kế tiếp. KHÔNG dừng để hỏi xác nhận vặt.

CHỈ DỪNG ở fork thật sự: §9 "Câu hỏi mở" của FR chưa chốt, hoặc spec mâu thuẫn, hoặc cần secret/tài nguyên ngoài (residential proxy, API key sàn, tài khoản gateway). Khi dừng, nêu rõ quyết định cần và đề xuất phương án.

PHẠM VI ĐỢT NÀY: build trọn Layer 0, rồi Layer 1 theo IMPLEMENTATION-ORDER. Critical path MVP: FR-INFRA-002 -> FR-PRICE-001 -> FR-PRICE-002 -> FR-DEAL-001/003 -> FR-WEB-003; song song bắt đầu sớm FR-SCRAPE-001/002/005 và FR-EXT-001/002/003. Sau khi xong mỗi layer, dừng lại tóm tắt các FR đã done + bằng chứng gate xanh để tôi review.

RÀNG BUỘC: Go 1.22 backend; Postgres 16 + TimescaleDB cho price; TypeScript + Chrome MV3; Next.js web; Python (Prophet/LightGBM) cho ML. Mọi tiền tệ BIGINT VND. Không commit secret (dùng biến môi trường/Vault như INFRA-003). Văn bản và commit tiếng Việt được; code và identifier tiếng Anh.

BẮT ĐẦU NGAY: đọc 5 tài liệu trên, in ra kế hoạch build Layer 0 (4 FR, kèm new_files và thứ tự), rồi thực thi FR đầu tiên.
```

## Prompt tiếp nối (các phiên sau)

```
Tiếp tục theo docs/feature-requests/IMPLEMENTATION-ORDER.md. Đọc lại SHIP-GUIDE.md (9 bất biến) và DATA-MODEL.md trước. Chọn FR khả thi kế tiếp (layer thấp nhất chưa done, mọi depends_on đã done, MUST trước). Lật status -> implementing trong BACKLOG, build theo new_files + sub_tasks, viết test §5, đối chiếu AC §4, chạy gate (go build/vet/test hoặc tsc/pytest) đến khi xanh, commit "FR-XXX: ...", status -> done. Lặp đến hết layer rồi dừng tóm tắt. Chỉ dừng giữa chừng ở §9 câu hỏi mở hoặc khi cần secret/tài nguyên ngoài.
```

## Ghi chú

- Nếu đã cài AGENTS.md của cyberos vào gốc repo thì memory tự kích hoạt; nếu chưa, prompt vẫn chạy đúng vì conventions lấy từ SHIP-GUIDE.md.
- Một FR là một đơn vị commit. Giữ BACKLOG.md đồng bộ status sau mỗi bước để phiên sau (hoặc agent khác) biết đang ở đâu.
- Khi agent dừng ở một "câu hỏi mở" (§9 của FR), đó là quyết định của anh - trả lời rồi cho chạy tiếp.
