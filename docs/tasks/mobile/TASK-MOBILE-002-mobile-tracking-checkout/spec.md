---
id: TASK-MOBILE-002
title: "Mobile theo dõi giá + nhận alert + universal checkout assistant - màn theo dõi/wishlist + biểu đồ giá + nhận push alert; checkout assistant gọi optimizer voucher (TASK-CART-003) hiển thị tổ hợp tối ưu KHI người dùng chủ động yêu cầu"
module: MOBILE
priority: SHOULD
status: ready_to_review
verify: T
phase: P3
milestone: P3 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-MOBILE-001, TASK-CART-003, TASK-PRICE-003, TASK-TRACK-002, TASK-TRACK-003]
depends_on: [TASK-MOBILE-001, TASK-CART-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §6 mục 6 (Universal checkout assistant browser/mobile - tối ưu voucher khi thanh toán; Premium + affiliate user-initiated)"
  - "docs/... §3.5 (tối ưu giỏ/voucher/freeship, knapsack stacking), §4.2 (affiliate chỉ user-initiated), §3.1 (mobile client của BFF)"
source_decisions:
  - "DEC-MOBILE-10: màn theo dõi giá + wishlist + biểu đồ đọc qua API có sẵn (TASK-PRICE-003 price-history, TASK-TRACK-002 wishlist, TASK-TRACK-003 alert_rule) - mobile là client mỏng, KHÔNG tính toán giá/sale phía client"
  - "DEC-MOBILE-11: checkout assistant CHỈ chạy khi người dùng chủ động bấm (user-initiated) - phù hợp ranh giới compliance affiliate hậu-Honey (§4.2), KHÔNG tự động kích hoạt nền lúc thanh toán"
  - "DEC-MOBILE-12: optimizer voucher/giỏ chạy ở backend (TASK-CART-003) - mobile gửi giỏ (đã tối thiểu hóa) + voucher khả dụng, nhận về tổ hợp tối ưu để HIỂN THỊ; mobile KHÔNG tự áp voucher hay tự thanh toán"
  - "DEC-MOBILE-13: checkout assistant hiển thị gợi ý voucher + tổng tiết kiệm rồi để NGƯỜI DÙNG tự áp mã trong app sàn - KHÔNG auto-click/auto-apply (chống abuse, tuân Chrome/sàn policy tương tự web)"
  - "DEC-MOBILE-14: push alert nhận qua FCM (TASK-MOBILE-001) -> mở đúng màn sản phẩm (deep-link nội bộ); alert đã được engine backend (TASK-TRACK-004) quyết định, mobile chỉ hiển thị"
  - "DEC-MOBILE-15: dữ liệu giỏ gửi lên optimizer tối thiểu hóa (chỉ product_id/qty/giá hiển thị) - KHÔNG gửi cookie/token sàn, nhất quán nguyên tắc tối thiểu hóa dữ liệu của extension"

language: "React Native 0.74 (TypeScript); biểu đồ qua victory-native; gọi backend qua httpClient của TASK-MOBILE-001"
service: shopass/apps/mobile/
new_files:
  - apps/mobile/src/track/trackScreen.tsx
  - apps/mobile/src/track/priceChart.tsx
  - apps/mobile/src/checkout/checkoutAssistant.tsx
  - apps/mobile/src/checkout/optimizerClient.ts
  - apps/mobile/src/checkout/optimizerClient.test.ts
  - apps/mobile/src/track/trackClient.ts
  - apps/mobile/src/track/trackClient.test.ts
modified_files:
  - apps/mobile/src/app/RootNavigator.tsx           # thêm tab Theo dõi + màn Checkout assistant
allowed_tools:
  - file_read: apps/mobile/**
  - file_write: apps/mobile/**
  - bash: cd apps/mobile && npm test
disallowed_tools:
  - tính toán sale ảo/giá phía client thay vì đọc API backend (vi phạm DEC-MOBILE-10)
  - tự động kích hoạt checkout assistant nền hoặc auto-apply voucher (vi phạm DEC-MOBILE-11/13, vi phạm compliance affiliate)
  - gửi cookie/token sàn lên optimizer (vi phạm DEC-MOBILE-15)

effort_hours: 10
sub_tasks:
  - "1.5h: trackClient.ts - gọi TASK-TRACK-002 (wishlist) + TASK-PRICE-003 (price-history) + TASK-TRACK-003 (alert_rule)"
  - "2.0h: trackScreen.tsx - danh sách theo dõi/wishlist + trạng thái sale (đọc từ API) + điều hướng tới chi tiết"
  - "1.5h: priceChart.tsx - render biểu đồ lịch sử giá từ price-history (client mỏng, không tính toán)"
  - "2.0h: checkoutAssistant.tsx - màn user-initiated: thu giỏ tối thiểu hóa, gọi optimizer, hiển thị gợi ý voucher + tổng tiết kiệm"
  - "1.5h: optimizerClient.ts - POST giỏ + voucher khả dụng tới TASK-CART-003, nhận tổ hợp tối ưu (chỉ hiển thị)"
  - "1.0h: deep-link push -> mở màn sản phẩm đúng (tích hợp FCM message handler của TASK-MOBILE-001)"
  - "1.0h: optimizerClient.test.ts + trackClient.test.ts - payload tối thiểu hóa; không auto-apply; đọc API đúng"

risk_if_skipped: "TASK-MOBILE-002 là lớp giá trị người dùng của mobile app (§6 mục 6) - theo dõi giá, nhận alert, và universal checkout assistant là lý do người dùng mở app mỗi ngày. Không có nó thì scaffold TASK-MOBILE-001 chỉ là vỏ rỗng. Rủi ro compliance là điểm dễ sai nhất: nếu checkout assistant tự động kích hoạt nền hoặc auto-apply voucher thì lặp lại đúng vết xe Honey (§4.2) mà SănDeal dựng cả định vị đạo đức để tránh - có thể bị sàn C&D hoặc store gỡ app. Nếu mobile tự tính sale ảo phía client thay vì đọc API thì logic giá phân mảnh giữa web/mobile/backend, dễ lệch và khó kiểm. Nếu gửi cookie/token sàn lên optimizer thì phá nguyên tắc tối thiểu hóa dữ liệu vốn là cam kết niềm tin cốt lõi. Checkout assistant phải tuyệt đối user-initiated + chỉ-hiển-thị, để người dùng tự áp mã trong app sàn."
---

## §1 - Mô tả (BCP-14 normative)

App mobile **MUST** cung cấp màn theo dõi giá + wishlist + biểu đồ (đọc từ API backend), nhận push alert mở đúng màn sản phẩm, và universal checkout assistant chạy CHỈ khi người dùng chủ động yêu cầu, gọi optimizer backend để HIỂN THỊ tổ hợp voucher tối ưu. Hợp đồng:

1. **MUST** đọc danh sách theo dõi/wishlist từ TASK-TRACK-002, lịch sử giá từ TASK-PRICE-003, và quy tắc alert từ TASK-TRACK-003 (DEC-MOBILE-10); mobile là client mỏng. **MUST NOT** tính toán sale ảo hay giá phía client.
2. **MUST** hiển thị biểu đồ lịch sử giá bằng dữ liệu trả về từ `GET /v1/products/{id}/price-history` (TASK-PRICE-003); biểu đồ là tầng render, không tính toán giá.
3. **MUST** nhận push alert qua FCM (TASK-MOBILE-001); chạm vào thông báo mở đúng màn sản phẩm liên quan qua deep-link nội bộ (DEC-MOBILE-14). Alert do engine backend (TASK-TRACK-004) quyết định; mobile chỉ hiển thị.
4. **MUST** chạy checkout assistant CHỈ khi người dùng chủ động bấm (user-initiated) (DEC-MOBILE-11); **MUST NOT** tự động kích hoạt nền lúc thanh toán.
5. **MUST** chạy tối ưu voucher/giỏ ở backend (TASK-CART-003) (DEC-MOBILE-12): mobile gửi giỏ đã tối thiểu hóa + danh sách voucher khả dụng, nhận về tổ hợp tối ưu để hiển thị. Mobile **MUST NOT** tự áp voucher hay tự thanh toán.
6. **MUST** chỉ HIỂN THỊ gợi ý voucher + tổng tiết kiệm rồi để người dùng tự áp mã trong app sàn (DEC-MOBILE-13); **MUST NOT** auto-click/auto-apply.
7. **MUST** tối thiểu hóa dữ liệu giỏ gửi lên optimizer: chỉ `product_id`, `qty`, giá hiển thị (DEC-MOBILE-15); **MUST NOT** gửi cookie/token sàn.
8. **MUST** xử lý trạng thái rỗng/không đủ dữ liệu nhã nhặn: SKU chưa đủ lịch sử (TASK-DEAL-002 trả UNKNOWN) hiển thị "chưa đủ dữ liệu" thay vì kết luận sai; optimizer không tìm được tổ hợp tốt hơn -> nói rõ "không có voucher áp dụng".
9. **MUST** dùng `httpClient` của TASK-MOBILE-001 cho mọi lời gọi (kế thừa auth + auto-refresh + gateway).
10. **SHOULD** đo client telemetry tối thiểu (mở checkout assistant, số gợi ý hiển thị) ở mức ẩn danh, tôn trọng consent; **MUST NOT** log nội dung giỏ chi tiết.
11. **MUST** hiển thị disclosure rõ ràng khi gợi ý có yếu tố affiliate (nhất quán TASK-AFFIL-004), để minh bạch với người dùng.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao mobile là client mỏng, đọc API (DEC-MOBILE-10, §1 #1)? Logic sale ảo (median90/p10/trailing_min) và tối ưu voucher (knapsack stacking) là lõi phức tạp, đã đặc tả ở TASK-DEAL-001 và TASK-CART-003 phía backend. Nếu mobile tự tính lại thì có hai (hoặc ba, cùng web) bản logic dễ lệch nhau - người dùng thấy "sale xịn" trên web nhưng "sale ảo" trên mobile cho cùng sản phẩm. Một nguồn sự thật ở backend; client chỉ hiển thị.

Vì sao checkout assistant tuyệt đối user-initiated (DEC-MOBILE-11, §1 #4)? Đây là ranh giới compliance sống còn (§4.2). Honey mang tiếng vì tự động can thiệp nền vào luồng mua. SănDeal dựng cả định vị đạo đức để khác Honey. Checkout assistant chỉ chạy khi người dùng chủ động bấm "tối ưu voucher" - không bao giờ tự bật. Vi phạm điều này có thể khiến sàn C&D hoặc store gỡ app.

Vì sao chỉ hiển thị, người dùng tự áp mã (DEC-MOBILE-13, §1 #6)? Auto-apply/auto-click voucher trong app sàn là hành vi tự động hóa dễ bị coi là abuse và vi phạm policy sàn (tương tự ràng buộc trên web ở TASK-CART-005). Mobile hiển thị "dùng mã X tiết kiệm Y" và để người dùng tự nhập trong app sàn - giữ con người trong vòng lặp, tránh tự động hóa rủi ro.

Vì sao optimizer ở backend, mobile gửi giỏ tối thiểu hóa (DEC-MOBILE-12/15, §1 #5, #7)? Tối ưu voucher cần dữ liệu catalog voucher + luật stacking per-country (TASK-CART-003/004) sống ở backend. Mobile chỉ biết giỏ hiện tại. Gửi giỏ tối thiểu hóa (product_id/qty/giá, KHÔNG cookie/token) lên backend tính rồi nhận kết quả - vừa đúng nơi có dữ liệu, vừa giữ nguyên tắc tối thiểu hóa dữ liệu (không bao giờ gửi token sàn).

Vì sao xử lý trạng thái rỗng nhã nhặn (§1 #8)? Cold-start nghĩa là nhiều SKU chưa đủ 90 ngày lịch sử (TASK-DEAL-002 trả UNKNOWN). Nếu mobile cố kết luận "sale xịn/ảo" trên dữ liệu thiếu thì sai và mất niềm tin. Hiển thị trung thực "chưa đủ dữ liệu" tốt hơn một kết luận bịa.

---

## §3 - Hợp đồng API / mã

### Optimizer client - gửi giỏ tối thiểu hóa (TypeScript)

```ts
// apps/mobile/src/checkout/optimizerClient.ts
import { http } from '../api/httpClient';

// CHỈ gửi dữ liệu tối thiểu hóa: product_id, qty, giá hiển thị. KHÔNG cookie/token sàn.
type CartLine = { product_id: number; qty: number; unit_price: number };
type OptimizeReq = { platform_id: number; lines: CartLine[]; voucher_codes: string[] };

type OptimizeResult = {
  suggested_vouchers: { code: string; saves: number }[];
  total_saving: number;
  has_affiliate: boolean; // để hiển thị disclosure
};

// gọi optimizer backend (TASK-CART-003); chỉ HIỂN THỊ kết quả, KHÔNG tự áp.
export async function optimizeCart(req: OptimizeReq): Promise<OptimizeResult> {
  const res = await http.post('/v1/cart/optimize', req);
  return res.json();
}
```

### Checkout assistant - user-initiated, chỉ hiển thị (TypeScript, rút gọn)

```tsx
// apps/mobile/src/checkout/checkoutAssistant.tsx (rút gọn)
export function CheckoutAssistant({ cart }: { cart: CartLine[] }) {
  const [result, setResult] = useState<OptimizeResult | null>(null);

  // CHỈ chạy khi người dùng chủ động bấm - KHÔNG useEffect tự động.
  async function onOptimizePressed() {
    const r = await optimizeCart({
      platform_id: cart.platformId,
      lines: cart.map(minimize),            // tối thiểu hóa: bỏ mọi field nhạy cảm
      voucher_codes: cart.availableVouchers,
    });
    setResult(r);
  }

  return (
    <View>
      <Button title="Tối ưu voucher" onPress={onOptimizePressed} />
      {result && (
        <>
          {result.has_affiliate && <DisclosureBanner />}  {/* minh bạch affiliate */}
          <SavingSummary total={result.total_saving} vouchers={result.suggested_vouchers} />
          <Text>Bạn tự áp các mã trên trong app sàn để nhận ưu đãi.</Text>
        </>
      )}
    </View>
  );
}
```

### Theo dõi giá - client mỏng đọc API (TypeScript)

```ts
// apps/mobile/src/track/trackClient.ts
export async function getPriceHistory(productId: number, range = '90d') {
  const res = await http.get(`/v1/products/${productId}/price-history?range=${range}`);
  return res.json(); // backend đã tính daily aggregate - mobile chỉ render
}

export async function getWishlist() {
  const res = await http.get('/v1/wishlist');
  return res.json();
}
```

---

## §4 - Acceptance criteria

1. Màn theo dõi hiển thị wishlist từ `GET /v1/wishlist`; không có tính toán giá nào phía client.
2. Biểu đồ giá render từ `price-history`; SKU chưa đủ lịch sử (UNKNOWN) hiển thị "chưa đủ dữ liệu" thay vì kết luận.
3. Push alert tới -> chạm mở đúng màn sản phẩm qua deep-link nội bộ.
4. Checkout assistant KHÔNG tự chạy: không có `useEffect`/auto-trigger gọi optimizer; chỉ chạy khi bấm nút.
5. Bấm "Tối ưu voucher" -> `POST /v1/cart/optimize` với payload chỉ chứa `product_id/qty/unit_price` + `voucher_codes`; KHÔNG có trường cookie/token.
6. Kết quả optimizer chỉ hiển thị (gợi ý mã + tổng tiết kiệm); KHÔNG có auto-apply/auto-click.
7. Optimizer không tìm được tổ hợp tốt hơn -> hiển thị "không có voucher áp dụng", không lỗi.
8. Kết quả có yếu tố affiliate -> hiển thị disclosure banner.
9. Mọi lời gọi đi qua `httpClient` của TASK-MOBILE-001 (kế thừa auth + auto-refresh).
10. Telemetry KHÔNG ghi nội dung giỏ chi tiết (review + test).

---

## §5 - Kiểm thử (verification)

```ts
// apps/mobile/src/checkout/optimizerClient.test.ts
test('payload optimizer chỉ chứa trường tối thiểu hóa, không cookie/token', async () => {
  const postSpy = jest.spyOn(http, 'post').mockResolvedValue(jsonRes({ total_saving: 0 }));
  await optimizeCart({
    platform_id: 1,
    lines: [{ product_id: 90112, qty: 2, unit_price: 89000 }],
    voucher_codes: ['FREESHIP'],
  });
  const body = postSpy.mock.calls[0][1] as any;
  const keys = Object.keys(body.lines[0]);
  expect(keys.sort()).toEqual(['product_id', 'qty', 'unit_price']);
  expect(JSON.stringify(body)).not.toContain('cookie');
  expect(JSON.stringify(body)).not.toContain('token');
});

test('optimizer không tìm tổ hợp -> hiển thị không có voucher', async () => {
  jest.spyOn(http, 'post').mockResolvedValue(jsonRes({ suggested_vouchers: [], total_saving: 0 }));
  const r = await optimizeCart(reqWithCart());
  expect(r.suggested_vouchers).toHaveLength(0);
});

// apps/mobile/src/checkout/checkoutAssistant.test.tsx (mô tả hành vi)
test('checkout assistant KHÔNG tự chạy khi mount', () => {
  const postSpy = jest.spyOn(http, 'post');
  render(<CheckoutAssistant cart={sampleCart} />);
  expect(postSpy).not.toHaveBeenCalled(); // chỉ chạy khi bấm
});

test('bấm Tối ưu voucher -> gọi optimizer một lần', async () => {
  const postSpy = jest.spyOn(http, 'post').mockResolvedValue(jsonRes({ total_saving: 110000 }));
  const { getByText } = render(<CheckoutAssistant cart={sampleCart} />);
  fireEvent.press(getByText('Tối ưu voucher'));
  await waitFor(() => expect(postSpy).toHaveBeenCalledTimes(1));
});

test('kết quả có affiliate -> hiển thị disclosure', async () => {
  jest.spyOn(http, 'post').mockResolvedValue(jsonRes({ total_saving: 50000, has_affiliate: true }));
  const { getByText, findByTestId } = render(<CheckoutAssistant cart={sampleCart} />);
  fireEvent.press(getByText('Tối ưu voucher'));
  expect(await findByTestId('disclosure-banner')).toBeTruthy();
});

// apps/mobile/src/track/trackClient.test.ts
test('biểu đồ đọc price-history từ API, không tính toán client', async () => {
  const getSpy = jest.spyOn(http, 'get').mockResolvedValue(jsonRes({ cells: [] }));
  await getPriceHistory(90112, '90d');
  expect(getSpy).toHaveBeenCalledWith('/v1/products/90112/price-history?range=90d');
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: trackClient.ts (đọc wishlist/price-history/alert) -> trackScreen.tsx + priceChart.tsx -> optimizerClient.ts -> checkoutAssistant.tsx -> deep-link handler -> tests. Tất cả dựng trên `httpClient` của TASK-MOBILE-001. Backend cần endpoint `POST /v1/cart/optimize` (mặt tiền của TASK-CART-003) nhận giỏ tối thiểu hóa. Deep-link push tái dùng FCM message handler của TASK-MOBILE-001, map `data.product_id` -> màn chi tiết. Test qua Jest + React Native Testing Library với mock http.

---

## §7 - Phụ thuộc

- TASK-MOBILE-001 - scaffold + `httpClient` (auth/auto-refresh) + FCM handler; mọi thứ dựng lên trên.
- TASK-CART-003 - optimizer voucher/giỏ ở backend; mobile gửi giỏ tối thiểu hóa, nhận tổ hợp tối ưu.
- TASK-PRICE-003 - API lịch sử giá cho biểu đồ.
- TASK-TRACK-002 / TASK-TRACK-003 - wishlist + alert_rule.
- TASK-TRACK-004 / TASK-DEAL-002 (liên quan) - engine alert + xử lý cold-start UNKNOWN.
- TASK-AFFIL-004 (liên quan) - chuẩn disclosure affiliate.
- Lib: `victory-native` (biểu đồ), React Navigation.

---

## §8 - Payload ví dụ

### Gọi optimizer (POST /v1/cart/optimize) - giỏ tối thiểu hóa

```json
{
  "platform_id": 1,
  "lines": [
    { "product_id": 90112, "qty": 2, "unit_price": 89000 },
    { "product_id": 90455, "qty": 1, "unit_price": 250000 }
  ],
  "voucher_codes": ["SHOPA30K", "PLATFORM50K", "FREESHIP"]
}
```

### Kết quả optimizer (hiển thị, không tự áp)

```json
{
  "suggested_vouchers": [
    { "code": "SHOPA30K", "saves": 30000 },
    { "code": "PLATFORM50K", "saves": 50000 },
    { "code": "FREESHIP", "saves": 30000 }
  ],
  "total_saving": 110000,
  "has_affiliate": false
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Thu giỏ tự động từ app sàn trên mobile (nếu có API chia sẻ) - hiện người dùng nhập/chọn sản phẩm thủ công vào assistant; tự động hóa cân nhắc sau, vẫn phải user-initiated.
- So sánh giá chéo sàn ngay trong assistant (TASK-PRICE-004) - thêm khi nhợp UX mobile.
- Widget màn hình chính (home screen widget) hiển thị giá theo dõi - tính năng nền tảng giai đoạn sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Checkout assistant tự chạy nền | test mount không gọi optimizer | lặp vết Honey, rủi ro C&D/gỡ app | Chỉ chạy khi bấm (DEC-MOBILE-11) |
| Auto-apply voucher trong app sàn | review + DEC-MOBILE-13 | abuse, vi phạm policy sàn | Chỉ hiển thị, người dùng tự áp |
| Gửi cookie/token sàn lên optimizer | test payload tối thiểu hóa | phá tối thiểu hóa dữ liệu | Chỉ product_id/qty/giá (DEC-MOBILE-15) |
| Tính sale ảo phía client | review + DEC-MOBILE-10 | logic lệch web/backend | Client mỏng đọc API |
| Kết luận trên SKU thiếu lịch sử | AC #2 | sai + mất niềm tin | Hiển thị "chưa đủ dữ liệu" (UNKNOWN) |
| Optimizer rỗng coi là lỗi | TestOptimizer rỗng | UX vỡ | Hiển thị "không có voucher áp dụng" |
| Thiếu disclosure affiliate | test disclosure banner | thiếu minh bạch | Banner khi has_affiliate (TASK-AFFIL-004) |
| Telemetry ghi nội dung giỏ | review telemetry | rò dữ liệu cá nhân | Chỉ đo ẩn danh, không nội dung giỏ |

---

## §11 - Ghi chú

- Ranh giới compliance sống còn: checkout assistant tuyệt đối user-initiated + chỉ-hiển-thị, người dùng tự áp mã trong app sàn (khác Honey).
- Mobile là client mỏng: logic sale ảo + tối ưu voucher ở backend (một nguồn sự thật), mobile chỉ render.
- Tối thiểu hóa dữ liệu giữ nguyên trên mobile: gửi giỏ chỉ product_id/qty/giá, không bao giờ cookie/token sàn.
- Trạng thái rỗng (cold-start UNKNOWN, optimizer không có tổ hợp) hiển thị trung thực, không bịa kết luận.
- Disclosure affiliate hiển thị rõ để minh bạch, nhất quán cam kết hậu-Honey.
- Dựng hoàn toàn trên httpClient của TASK-MOBILE-001 nên kế thừa auth + auto-refresh + gateway sẵn.

---

*Hết TASK-MOBILE-002. Status: ready_to_review (awaiting HITL) (mục tiêu audit 10/10).*
