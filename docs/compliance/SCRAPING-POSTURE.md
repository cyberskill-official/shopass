# Tư thế pháp lý thu thập dữ liệu công khai (scraping posture)

**Sản phẩm:** Shopass (thương hiệu công khai Shopass / CyberSkill)  
**Miền:** `shopass.cyberskill.world`, `api.shopass.cyberskill.world`  
**Phiên bản:** R37 draft — chờ rà soát nội bộ / tư vấn (nếu có)  
**Chủ sở hữu tài liệu:** Stephen Cheng (Founder)  
**Liên hệ pháp lý / takedown:** [info@cyberskill.world](mailto:info@cyberskill.world)

---

## 1. Tóm tắt điều hành

Shopass giúp người dùng **theo dõi giá sản phẩm** trên Shopee, TikTok Shop và Lazada (Việt Nam). Backend `scrapesvc` đọc **thông tin listing công khai** (giá, tồn kho, tín hiệu khuyến mãi) cho các SKU mà **người dùng đã chủ động thêm vào danh sách theo dõi** — không thu thập dữ liệu cá nhân từ sàn, không dùng phiên đăng nhập của người dùng, không vượt tường xác thực để lấy dữ liệu riêng tư.

Chúng tôi **không** coi đây là “quyền tuyệt đối” để thu thập. Điều khoản sử dụng (ToS) của từng sàn **hạn chế truy cập tự động**; memo này ghi nhận ma sát đó một cách trung thực và mô tả cách Shopass giảm thiểu rủi ro kỹ thuật và pháp lý.

**Phạm vi memo:** phân tích ToS theo sàn; phạm vi dữ liệu công khai; robots / rate budget; tối thiểu hóa & lưu trữ; quy trình takedown; quan hệ chương trình affiliate (R29); FAQ báo chí.

---

## 2. Khung định vị sản phẩm

| Khía cạnh | Cam kết Shopass |
|---|---|
| **Kích hoạt** | Chỉ quét SKU đã được user (hoặc luồng onboarding) **chủ động** thêm qua `POST /v1/track` |
| **Nguồn dữ liệu** | Trang / endpoint listing **không đăng nhập** (`is_login:false` trên Shopee; farm Playwright cho TikTok/Lazada khi cần) |
| **Không thu thập** | Cookie phiên user, mật khẩu, token, hồ sơ buyer/seller, tin nhắn, địa chỉ, lịch sử đơn hàng |
| **Mục đích** | `price_tracking` — xem biểu đồ giá, cảnh báo đáy, phát hiện sale ảo (xem `docs/compliance/DATA-INVENTORY.md`) |
| **Affiliate** | Liên kết affiliate chỉ khi user **chủ động bấm**; scraping backend **không** gắn cookie affiliate (R29 chưa live — xem §7) |

Shopass **không** positioning là công cụ “cào toàn sàn”, không bán lại catalog nguyên vẹn, không reverse-engineer API ký (ví dụ TikTok `msToken` / `_signature`).

---

## 3. Phân tích ToS theo sàn (thẳng thắn)

> **Lưu ý:** Đây là đánh giá nội bộ dựa trên điều khoản công khai và thực hành kỹ thuật hiện tại. **Không phải ý kiến pháp lý.** Không trích dẫn án lệ giả định.

### 3.1 Shopee Việt Nam

**Điều khoản liên quan (tóm lược):** Shopee thường cấm robot, truy vấn tự động, scraping / trích xuất dữ liệu hàng loạt ngoài API chính thức, và sử dụng công cụ tự động làm suy giảm hệ thống.

**Ma sát với Shopass:** Cao. Dù chỉ đọc giá listing công khai, nhịp quét định kỳ của `scrapesvc` vẫn có thể bị xếp vào “automated access” theo nghĩa rộng của ToS.

**Cách Shopass định vị:**
- Chỉ lấy trường giá / tồn kho / khuyến mãi từ listing user đã chọn — tương đương việc user tự mở trang sản phẩm nhiều lần để so giá.
- Adapter Shopee ưu tiên endpoint PDP công khai; **không** dùng cookie phiên người dùng.
- Khi Shopee chặn (403, challenge, CAPTCHA): **dừng / giảm tần suất**, không leo thang bypass.

**Rủi ro còn lại:** Shopee có thể gửi thông báo chấm dứt, chặn IP/proxy, hoặc liên hệ affiliate nếu có hợp đồng. Shopass duy trì inbox takedown (§6).

### 3.2 TikTok Shop Việt Nam

**Điều khoản liên quan (tóm lược):** ByteDance / TikTok Shop hạn chế truy cập tự động, scraping, và tái sử dụng dữ liệu ngoài kênh được phép; WAF và chữ ký request nghiêm.

**Ma sát với Shopass:** Cao đến rất cao. TikTok Shop buộc render DOM; adapter **không** tái tạo chữ ký API nội bộ — chỉ đọc trang đã render trong farm Playwright (DEC-SCRAPE-28/29).

**Cách Shopass định vị:**
- Phạm vi hẹp: giá, `list_price`, tín hiệu flash sale, sold/stock nếu hiển thị công khai trên PDP.
- Proxy residential tier enterprise; pacing chậm hơn Shopee.
- Fail-closed: không ghi giá fabricate khi extract thất bại.

**Rủi ro còn lại:** TikTok có biện pháp chống bot mạnh; tần suất quét có thể phải giảm mạnh hoặc tạm dừng theo sàn.

### 3.3 Lazada Việt Nam

**Điều khoản liên quan (tóm lược):** Lazada (Alibaba ecosystem) cấm truy cập tự động trái phép, scraping hàng loạt, và can thiệp WAF/Akamai.

**Ma sát với Shopass:** Cao. Lazada dùng Akamai; adapter farm đọc embedded JSON / DOM trên PDP công khai.

**Cách Shopass định vị:** Giống Shopee — SKU user-initiated, không auth wall, residential proxy + pacing, throttle khi health monitor báo degraded/broken.

**Rủi ro còn lại:** Chặn IP, CAPTCHA slider, hoặc thông báo pháp lý từ Lazada / đối tác.

### 3.4 Tổng hợp ma sát

| Sàn | Ma sát ToS | Dữ liệu lấy | Giảm thiểu chính |
|---|---|---|---|
| Shopee VN | Cao | Giá listing công khai | User-initiated SKU, no login, pacing, deny khi block |
| TikTok Shop VN | Rất cao | Giá PDP render công khai | DOM-only, no signed API, enterprise proxy, fail-closed |
| Lazada VN | Cao | Giá PDP công khai | Farm + Akamai-aware TLS, pacing, throttle |

**Thành thật:** Không có cách diễn đạt nào biến quét định kỳ thành “được ToS chấp thuận rõ ràng”. Chiến lược Shopass là **phạm vi hẹp, minh bạch, sẵn sàng dừng / điều chỉnh** khi sàn phản đối.

---

## 4. Phạm vi dữ liệu công khai

Căn cứ `price_snapshot` và `docs/compliance/DATA-INVENTORY.md`:

| Trường | Thu thập? | Ghi chú |
|---|---|---|
| `price` | Có | Giá bán hiện tại (BIGINT VND) |
| `list_price` | Có | Giá gốc / niêm yết nếu hiển thị |
| `stock` | Có | Tín hiệu còn hàng / hết hàng (boolean hoặc mức coarse) |
| `sold` | Có | Số đã bán nếu sàn hiển thị công khai trên PDP |
| `flash_sale` | Có | Cờ khuyến mãi / flash sale |
| Tên shop / title listing | Có (metadata SKU) | Phục vụ hiển thị & canonical match — không phải PII buyer |
| Email, SĐT, địa chỉ buyer/seller | **Không** | Ngoài phạm vi |
| Cookie, token, session | **Không** | Extension không gửi lên server; scraper không dùng phiên user |
| Giỏ hàng / voucher / lịch sử mua | **Không** (scraping) | Extension đọc cart chỉ trên client, có consent `cart_read` — tách biệt pipeline scrape |

Dữ liệu scrape **không** nằm trong bảng personal-data inventory trừ khi join qua `user_tracked_product` (quan hệ user ↔ SKU do user tạo).

---

## 5. Robots.txt, proxy residential và rate budget

### 5.1 Stance về robots.txt

- Shopass **theo dõi** robots.txt / meta robots của từng sàn khi vận hành (mục tiêu R24: ghi nhận trong cấu hình `region/` per-platform).
- **Không** coi robots.txt là giấy phép pháp lý; coi là **tín hiệu vận hành**: nếu path listing bị `Disallow` rõ ràng cho bot, **không quét path đó** cho đến khi có rà soát lại.
- Trang marketing Shopass (`shopass.cyberskill.world`) có `robots.ts` riêng — không liên quan pipeline scrape.

### 5.2 Chính sách kỹ thuật (đã có trong `services/scrape`)

| Lớp | Triển khai | Tham chiếu |
|---|---|---|
| **Residential proxy** | Mặc định cho Shopee / TikTok / Lazada; datacenter chỉ khi xác nhận không WAF | `services/scrape/internal/proxy/` |
| **Pacing + jitter** | Delay ngẫu nhiên `[min,max]` per-platform trước mỗi request | `services/scrape/internal/pacing/limiter.go` |
| **Concurrency cap** | Giới hạn song song per-platform, độc lập với pacing | orchestrator pool |
| **Cost guard** | Trần ngân sách proxy/ngày | `proxy/costguard.go` |
| **Health monitor** | Degraded → 50% tần suất; Broken → ~10% probe | `health/monitor.go` `ShouldThrottle` |
| **Deny-by-default khi block** | CAPTCHA / hard block → không ghi snapshot fabricate; job backoff | adapters fail-closed; TikTok adapter |
| **CAPTCHA** | Phát hiện → hàng đợi thủ công (solver giả lập hiện tại); **R25** sẽ thêm manual queue pluggable | `captcha/detect.go`, backlog R25 |

**Nguyên tắc:** Khi sàn từ chối truy cập, Shopass **giảm hoặc dừng**, không leo thang (không vòng lặp bypass CAPTCHA tự động ở production cho đến khi R25 được phê duyệt và ghi rõ trong policy).

### 5.3 Rate budget (định hướng)

Tần suất quét theo tier (`hot` ≤5 phút, `warm` 1–6 giờ, `cold` 24 giờ) — chỉ áp dụng cho SKU đã tracked, không quét toàn catalog. Budget cụ thể per-platform sẽ được tinh chỉnh sau R24 battle-test; memo này ghi **intent**: “đủ chậm để không gây hại hạ tầng, đủ nhanh để alert có giá trị”.

---

## 6. Tối thiểu hóa dữ liệu và lưu trữ

- **Delta-only:** Chỉ ghi `price_snapshot` khi `(price, list_price, stock, flash_sale)` thay đổi — giảm lưu trữ và tránh lưu bản sao listing không cần thiết.
- **Retention (mặc định sản phẩm, chờ R19 / counsel):**
  - Raw `price_snapshot`: **18 tháng** (TimescaleDB retention policy).
  - Aggregate `price_daily`: **không giới hạn** (phục vụ forecast / biểu đồ dài hạn).
  - Nén chunk sau **30 ngày**.
- **Erasure:** Xóa tài khoản → hard-delete `user_tracked_product` và quan hệ user; snapshot sản phẩm shared có thể giữ nếu không còn join PII (xem DATA-INVENTORY erasure summary).
- **B2B / API:** Không expose raw snapshot qua API công khai; trend B2B chỉ từ aggregate ẩn danh (DEC-B2B-33).

Chi tiết inventory: `docs/compliance/DATA-INVENTORY.md`. Quyết định retention production: backlog **R19**.

---

## 7. Quy trình phản hồi takedown / thông báo pháp lý từ sàn

| Hạng mục | Giá trị |
|---|---|
| **Owner** | Stephen Cheng (Founder) |
| **Inbox** | [info@cyberskill.world](mailto:info@cyberskill.world) |
| **SLA acknowledge** | **48 giờ** làm việc kể từ khi nhận thông báo hợp lệ |
| **SLA phản hồi substantiative** | **7 ngày** làm việc (kế hoạch hành động: dừng SKU / sàn / toàn bộ scrape tùy yêu cầu) |

**Quy trình:**

1. **Tiếp nhận** — email tới inbox; ghi `received_at`, nguồn (sàn / luật sư / cơ quan), SKU hoặc phạm vi yêu cầu.
2. **Acknowledge (≤48h)** — xác nhận đã nhận, owner, timeline phản hồi đầy đủ.
3. **Đánh giá (ngày 1–3)** — xác minh tính hợp lệ; xem log scrape, SKU affected, ToS mapping; tư vấn counsel nếu cần.
4. **Hành động (ngày 3–7)** — ít nhất một trong: (a) ngừng quét SKU/sàn cụ thể; (b) giảm rate; (c) xóa dữ liệu lịch sử nếu yêu cầu hợp lý; (d) phản hồi bằng văn bản giải thích phạm vi và cam kết tuân thủ.
5. **Đóng & ghi sổ** — append evidence vào `docs/tasks/improvement/LEDGER.md`; cập nhật posture nếu thay đổi chính sách.

Thông báo vi phạm dữ liệu cá nhân (PDPL) dùng `docs/compliance/BREACH-RUNBOOK.md` — tách workflow khỏi takedown scraping.

---

## 8. Quan hệ chương trình affiliate (R29)

**Trạng thái:** R29 (affiliate programs live + attribution logging) **chưa live** — chưa đăng ký / chưa bật postback production.

**Posture (áp dụng trước khi R29 live):**

- Scraping giá listing và affiliate tracking là **hai pipeline tách biệt**. Scraper backend **không** set cookie affiliate, **không** auto-redirect checkout.
- Khi R29 live, Shopass dự kiến là **đối tác affiliate đăng ký chính thức** (Shopee Affiliate, TikTok Shop partner, Lazada affiliate nếu có). Trạng thái “registered partner” **khác** bot ẩn danh: có contact account manager, disclosure công khai, và kênh nhận thông báo từ network.
- Dù là partner, **ToS vẫn có thể cấm automated scraping** — affiliate status không tự động hợp thức hóa quét giá. Shopass vẫn tuân §5–§6 và sẵn sàng điều chỉnh theo yêu cầu network.
- Extension: affiliate link chỉ khi user **chủ động bấm** + disclosure (NFR-AFFIL-001).

---

## 9. FAQ báo chí / nền tảng (3 câu hỏi)

### Câu 1 — Shopass có “cào” Shopee/TikTok/Lazada không?

**Tiếng Việt:** Shopass theo dõi **giá công khai** trên trang sản phẩm cho các mục người dùng **tự chọn** — giống việc mở lại trang để xem giá nhiều lần, nhưng có biểu đồ và cảnh báo. Chúng tôi **không** lấy thông tin cá nhân, không dùng tài khoản người dùng, không bán lại toàn bộ dữ liệu sàn. Khi sàn chặn truy cập, chúng tôi **giảm hoặc dừng** thay vì cố bypass.

**English:** Shopass tracks **public listing prices** for products users **explicitly choose** to follow — similar to revisiting a product page for price checks, with charts and alerts. We do **not** collect personal data, use user login sessions, or resell full marketplace catalogs. When a platform blocks access, we **slow down or stop** rather than aggressively bypass defenses.

### Câu 2 — Việc này có vi phạm điều khoản sàn không?

**Tiếng Việt:** Các sàn thường **hạn chế truy cập tự động** trong ToS — chúng tôi không che giấu điều đó. Shopass giới hạn phạm vi (chỉ giá / tồn kho / khuyến mãi công khai), dùng proxy residential và pacing, và có quy trình phản hồi thông báo pháp lý trong 48h / 7 ngày. Chúng tôi khuyến nghị rà soát tư vấn pháp lý VN trước launch quy mô lớn.

**English:** Marketplace ToS often **restrict automated access** — we do not pretend otherwise. Shopass limits scope (public price/stock/promo signals only), uses residential proxies and pacing, and maintains a legal-notice response process (48h ack / 7d substantive reply). We recommend Vietnamese legal review before large-scale launch.

### Câu 3 — Shopass khác extension “ săn voucher / Honey ” thế nào?

**Tiếng Việt:** Shopass **không** thay cookie affiliate ngầm, **không** auto-applied voucher trái ý user, **không** đọc phiên đăng nhập sàn trên server. Affiliate (khi bật) chỉ qua **click có chủ ý**; minh bạch tại `/minh-bach`. Giá trị cốt lõi là **theo dõi giá và sale thật/sale ảo**, không phải chiếm hoa hồng creator.

**English:** Shopass does **not** silently swap affiliate cookies, force-apply vouchers, or pull marketplace login sessions server-side. Affiliate (when enabled) requires an **intentional click**; transparency at `/minh-bach`. Core value is **price tracking and fake-sale detection**, not taking creators’ commissions.

---

## 10. Stephen ask — quyết định cần làm

**Câu hỏi:** Có nên thuê **tư vấn pháp lý Việt Nam** rà soát memo này (và liên kết với `/dieu-khoan`, `/chinh-sach-bao-mat`) **trước launch / press cycle R53** không?

**Khuyến nghị của agent:** **Có** — engagement **nửa ngày** (~4 giờ billable) tập trung:
1. ToS scraping vs. price-comparison tools under VN law;
2. PDPL angle (public product data vs. user tracking joins);
3. Template phản hồi takedown / platform legal notice;
4. Whether affiliate registration (R29) changes recommended posture.

**Không commit counsel opinion vào repo** — chỉ cập nhật memo + waiver trong DPIA nếu counsel ký off.

**Action items cho Stephen:**
- [ ] Đọc memo này và FAQ §9
- [ ] Xác nhận `info@cyberskill.world` route tới inbox được monitor (48h SLA)
- [ ] Quyết định: engage VN counsel pre-launch? (recommended: yes, half-day)
- [ ] Nếu yes: chuyển memo + DATA-INVENTORY + NFR-AFFIL-001 cho counsel

---

*Tài liệu nội bộ / công khai có kiểm soát. Cập nhật lần cuối: 2026-07-26 (R37).*
