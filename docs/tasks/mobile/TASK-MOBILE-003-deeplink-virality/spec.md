---
id: TASK-MOBILE-003
title: "Deep-link + share-on-sale virality + referral - sinh share link mang referral_code (TASK-BILL-004) khi user chia sẻ deal đang sale; mở app từ link (universal/app link) gắn attribution lúc đăng ký; chống self-referral phía client"
module: MOBILE
priority: COULD
status: ready_to_review
verify: T
phase: P3
milestone: P3 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-MOBILE-001, TASK-BILL-004, TASK-MOBILE-002, TASK-TRUST-004]
depends_on: [TASK-MOBILE-001, TASK-BILL-004]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.7 (Virality: share-on-sale, referral code; SEA sequencing)"
  - "docs/... §6 (loyalty/virality), §3.4 (referral_code, app_user.referral_code_id), §5.3 (chống referral abuse/farming)"
source_decisions:
  - "DEC-MOBILE-20: share link là universal link (iOS) / app link (Android) trỏ tới một URL https://sandeal/... mang product_id + ref (referral_code của người chia sẻ); mở app nếu đã cài, mở web fallback nếu chưa"
  - "DEC-MOBILE-21: referral_code lấy từ TASK-BILL-004 (mỗi user một mã) - mobile KHÔNG tự sinh mã; gọi backend lấy mã của user hiện tại rồi gắn vào link"
  - "DEC-MOBILE-22: share-on-sale CHỈ tạo link khi người dùng chủ động bấm 'Chia sẻ' trên một deal - user-initiated, KHÔNG tự đăng/tự gửi link nền (chống spam, nhất quán seeding KHÔNG spam affiliate §5.7)"
  - "DEC-MOBILE-23: khi app mở từ deep-link có ref, LƯU ref tạm (pending) tới lúc người nhận đăng ký; attribution chỉ gắn vào app_user.referral_code_id tại đăng ký (TASK-BILL-004), bất biến sau đó"
  - "DEC-MOBILE-24: chặn self-referral phía client như lớp phòng vệ đầu (user đang đăng nhập không tự dùng ref của chính mình); backend (TASK-BILL-004) vẫn là người gác cổng cuối + phát sự kiện anti-fraud (TASK-TRUST-004)"
  - "DEC-MOBILE-25: deep-link payload chỉ mang product_id + ref (mã ngắn) - KHÔNG nhúng token/PII của người chia sẻ vào link công khai"

language: "React Native 0.74 (TypeScript); deep-link qua React Navigation linking + react-native-share"
service: shopass/apps/mobile/
new_files:
  - apps/mobile/src/share/shareLink.ts
  - apps/mobile/src/share/shareLink.test.ts
  - apps/mobile/src/deeplink/linkHandler.ts
  - apps/mobile/src/deeplink/linkHandler.test.ts
  - apps/mobile/src/deeplink/pendingReferral.ts
  - apps/mobile/src/share/ShareSheet.tsx
modified_files:
  - apps/mobile/src/auth/authClient.ts             # gắn pending ref vào luồng đăng ký (TASK-BILL-004)
allowed_tools:
  - file_read: apps/mobile/**
  - file_write: apps/mobile/**
  - bash: cd apps/mobile && npm test
disallowed_tools:
  - tự sinh referral_code phía mobile thay vì lấy từ TASK-BILL-004 (vi phạm DEC-MOBILE-21)
  - tự động đăng/gửi share link nền không qua hành động người dùng (vi phạm DEC-MOBILE-22, spam)
  - nhúng token/PII của người chia sẻ vào deep-link công khai (vi phạm DEC-MOBILE-25)

effort_hours: 6
sub_tasks:
  - "1.0h: shareLink.ts - lấy referral_code của user (TASK-BILL-004), dựng universal/app link mang product_id + ref"
  - "1.0h: ShareSheet.tsx - màn chia sẻ user-initiated (bấm 'Chia sẻ' trên deal) qua react-native-share"
  - "1.5h: linkHandler.ts - parse universal/app link, route tới màn sản phẩm, trích ref nếu có"
  - "1.0h: pendingReferral.ts - lưu ref tạm (pending) tới lúc đăng ký; chống self-referral phía client"
  - "1.0h: authClient.ts - gắn pending ref vào luồng đăng ký (gọi TASK-BILL-004 attribution)"
  - "0.5h: cấu hình apple-app-site-association + assetlinks.json (universal/app link)"
  - "1.0h: shareLink.test.ts + linkHandler.test.ts + pendingReferral.test.ts - link đúng định dạng; không token trong link; self-referral bị chặn; attribution gắn lúc đăng ký"

risk_if_skipped: "TASK-MOBILE-003 là động cơ tăng trưởng viral của mobile (§5.7: share-on-sale + referral). Không có nó thì app tăng trưởng chỉ bằng kênh trả phí, mất đòn bẩy network effect mà tài liệu coi là moat (§5.6). Rủi ro chính là biến virality thành spam hoặc gian lận: nếu mobile tự động đăng/gửi link nền thì lặp đúng kiểu spam affiliate mà §5.7 cấm (seeding cộng đồng KHÔNG spam link) - có thể bị cộng đồng tẩy chay và sàn để ý. Nếu để mobile tự sinh referral_code hay tự trả thưởng thì bỏ qua hook anti-fraud (TASK-TRUST-004, §5.3) và mở cửa farming tài khoản ảo. Nếu nhúng token/PII của người chia sẻ vào link công khai thì rò dữ liệu cá nhân ra ngoài. Virality phải user-initiated, mã lấy từ backend, attribution + chống gian lận do backend gác cổng - mobile chỉ là lớp phòng vệ đầu và bề mặt chia sẻ."
---

## §1 - Mô tả (BCP-14 normative)

App mobile **MUST** cho người dùng chia sẻ deal đang sale qua deep-link mang `referral_code` (lấy từ TASK-BILL-004), mở app từ universal/app link tới đúng màn sản phẩm, và gắn attribution lúc người nhận đăng ký. Hợp đồng:

1. **MUST** dựng share link dạng universal link (iOS) / app link (Android) trỏ tới URL `https://sandeal/...` mang `product_id` + `ref` (referral_code của người chia sẻ) (DEC-MOBILE-20); mở app nếu đã cài, fallback web nếu chưa.
2. **MUST** lấy `referral_code` từ TASK-BILL-004 (mỗi user một mã) (DEC-MOBILE-21); mobile **MUST NOT** tự sinh mã.
3. **MUST** chỉ tạo + chia sẻ link khi người dùng chủ động bấm "Chia sẻ" trên một deal (user-initiated) (DEC-MOBILE-22); **MUST NOT** tự động đăng/gửi link nền.
4. **MUST** khi mở app từ deep-link có `ref`: route tới đúng màn sản phẩm (`product_id`) và LƯU `ref` tạm (pending) (DEC-MOBILE-23).
5. **MUST** gắn attribution chỉ tại lúc người nhận đăng ký, qua `app_user.referral_code_id` của TASK-BILL-004 (DEC-MOBILE-23); attribution bất biến sau khi gắn; nếu người nhận đã có referrer thì **MUST NOT** ghi đè.
6. **MUST** chặn self-referral phía client như lớp phòng vệ đầu (DEC-MOBILE-24): người dùng đang đăng nhập không được dùng `ref` là mã của chính mình. Backend (TASK-BILL-004) vẫn là người gác cổng cuối.
7. **MUST** chỉ đưa `product_id` + `ref` (mã ngắn) vào deep-link công khai (DEC-MOBILE-25); **MUST NOT** nhúng token/PII của người chia sẻ.
8. **MUST** để việc trả thưởng referral cho backend (TASK-BILL-004 phát sự kiện cho TASK-TRUST-004 anti-fraud + delay reward); mobile **MUST NOT** tự trả thưởng hay tự tăng `uses`.
9. **MUST** xử lý deep-link không hợp lệ/thiếu trường nhã nhặn: link thiếu `product_id` -> mở màn chủ; `ref` sai định dạng -> bỏ qua ref, vẫn mở sản phẩm.
10. **SHOULD** đo telemetry tối thiểu (số lần bấm Chia sẻ, số lần mở từ deep-link) ở mức ẩn danh; **MUST NOT** ghi cặp (người chia sẻ -> người nhận) ngoài attribution chính thức ở backend.
11. **MUST** dùng `httpClient` của TASK-MOBILE-001 cho lời gọi lấy mã + đăng ký (kế thừa auth + gateway).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao universal/app link, không scheme riêng (DEC-MOBILE-20, §1 #1)? Scheme riêng (`sandeal://`) không mở được web fallback và dễ bị app khác chiếm. Universal link (iOS) / app link (Android) là URL https thật: nếu có app thì mở app, không có thì mở web - một link phục vụ cả người đã cài lẫn người mới (đúng mục tiêu viral: người nhận có thể chưa cài app).

Vì sao mã từ backend, không tự sinh (DEC-MOBILE-21, §1 #2)? `referral_code` là tài sản gắn anti-fraud (TASK-BILL-004/TRUST-004). Nếu mobile tự sinh thì có thể trùng, giả, hoặc bỏ qua kiểm tra. Một nguồn phát mã duy nhất ở backend đảm bảo mã unique + truy vết được + gắn đúng người.

Vì sao user-initiated, không tự đăng nền (DEC-MOBILE-22, §1 #3)? §5.7 nói rõ seeding cộng đồng KHÔNG được spam link affiliate nền. Tự động đăng/gửi link là spam - phá niềm tin cộng đồng và lặp đúng tiếng xấu Honey. Người dùng chủ động bấm "Chia sẻ" mới tạo link; SănDeal không bao giờ tự gửi thay họ.

Vì sao attribution chỉ gắn lúc đăng ký + bất biến (DEC-MOBILE-23, §1 #5)? Gắn referrer tại đăng ký là thời điểm tự nhiên + một-lần, khớp mô hình TASK-BILL-004 (referral_code_id bất biến). Lưu ref pending từ lúc click tới lúc đăng ký giữ được attribution qua khoảng người dùng cài app rồi mới tạo tài khoản. Bất biến tránh tranh chấp "ai giới thiệu" về sau.

Vì sao chặn self-referral phía client nhưng backend vẫn gác cổng (DEC-MOBILE-24, §1 #6)? Chặn phía client là trải nghiệm tốt (báo ngay "không thể tự giới thiệu") + giảm tải. Nhưng client không đáng tin tuyệt đối (có thể bị sửa). TASK-BILL-004 đã cấm self-referral ở backend (DEC-BILL-18) - đó là hàng rào thật. Mobile là lớp phòng vệ đầu, không phải lớp duy nhất.

Vì sao không nhúng token/PII vào link (DEC-MOBILE-25, §1 #7)? Share link đi ra ngoài (gửi bạn bè, đăng nhóm) - là dữ liệu công khai. Nhúng token nghĩa là rò phiên; nhúng PII (email/tên) nghĩa là rò danh tính người chia sẻ. Chỉ `product_id` + mã ref ngắn (vốn không phải bí mật) là an toàn để công khai.

Vì sao trả thưởng do backend + delay (DEC-MOBILE-24, §1 #8)? §5.3 cảnh báo referral abuse + farming tài khoản ảo. Trả thưởng ngay phía client là mời gọi gian lận. TASK-BILL-004 phát sự kiện cho TASK-TRUST-004 (anti-fraud) và delay reward tới khi vượt kiểm tra. Mobile không chạm vào tiền thưởng - chỉ tạo link + gắn attribution.

---

## §3 - Hợp đồng API / mã

### Dựng share link - mã từ backend, không token (TypeScript)

```ts
// apps/mobile/src/share/shareLink.ts
import { http } from '../api/httpClient';

const BASE = 'https://sandeal.app/p';

// lấy referral_code của user hiện tại (TASK-BILL-004) - KHÔNG tự sinh.
async function myReferralCode(): Promise<string> {
  const res = await http.get('/v1/referral/code'); // trả mã của user đăng nhập
  const { code } = await res.json();
  return code;
}

// link công khai chỉ mang product_id + ref (mã ngắn). KHÔNG token/PII.
export async function buildShareLink(productId: number): Promise<string> {
  const ref = await myReferralCode();
  return `${BASE}/${productId}?ref=${encodeURIComponent(ref)}`;
}
```

### Xử lý deep-link + lưu ref pending (TypeScript)

```ts
// apps/mobile/src/deeplink/linkHandler.ts
import { savePendingReferral } from './pendingReferral';

type ParsedLink = { productId?: number; ref?: string };

export function parseDeepLink(url: string): ParsedLink {
  const u = new URL(url);
  const m = u.pathname.match(/\/p\/(\d+)/);
  return { productId: m ? Number(m[1]) : undefined, ref: u.searchParams.get('ref') ?? undefined };
}

// mở app từ link: route tới sản phẩm; lưu ref pending (gắn sau, lúc đăng ký).
export async function handleDeepLink(url: string, nav: Navigator): Promise<void> {
  const { productId, ref } = parseDeepLink(url);
  if (ref) {
    await savePendingReferral(ref); // pending tới lúc đăng ký
  }
  if (productId) {
    nav.navigate('Product', { productId });
  } else {
    nav.navigate('Home'); // link thiếu product -> màn chủ
  }
}
```

### Chống self-referral phía client + attribution lúc đăng ký (TypeScript)

```ts
// apps/mobile/src/deeplink/pendingReferral.ts
import { getMyCode } from '../share/shareLink';

// lớp phòng vệ đầu: người đăng nhập không tự dùng mã của chính mình.
export async function applyPendingReferralAtSignup(): Promise<string | null> {
  const ref = await readPendingReferral();
  if (!ref) return null;
  const own = await currentUserCodeOrNull(); // null nếu chưa đăng nhập
  if (own && own === ref) {
    await clearPendingReferral();
    return null; // self-referral - backend cũng sẽ từ chối (TASK-BILL-004)
  }
  // attribution thật do TASK-BILL-004 ghi vào app_user.referral_code_id lúc đăng ký
  return ref;
}
```

---

## §4 - Acceptance criteria

1. Bấm "Chia sẻ" trên một deal -> sinh universal/app link `https://sandeal.app/p/{productId}?ref={code}`.
2. `ref` trong link là `referral_code` lấy từ `GET /v1/referral/code` (TASK-BILL-004); mobile KHÔNG tự sinh mã.
3. Link KHÔNG chứa token hay PII (email/tên) của người chia sẻ (kiểm bằng test).
4. Share chỉ xảy ra khi người dùng bấm; KHÔNG có đường tự động đăng/gửi link nền (review + test).
5. Mở app từ link có `ref` -> route tới đúng màn sản phẩm + lưu `ref` pending.
6. Người nhận đăng ký -> attribution gắn qua luồng TASK-BILL-004; nếu đã có referrer -> KHÔNG ghi đè.
7. Người dùng đang đăng nhập dùng `ref` là mã của chính mình -> bị chặn phía client (và backend từ chối).
8. Link thiếu `product_id` -> mở màn chủ; `ref` sai định dạng -> bỏ qua ref, vẫn mở sản phẩm.
9. Mobile KHÔNG tự trả thưởng hay tự tăng `uses`; việc đó do backend (TASK-BILL-004 -> TASK-TRUST-004).
10. Mọi lời gọi đi qua `httpClient` của TASK-MOBILE-001.

---

## §5 - Kiểm thử (verification)

```ts
// apps/mobile/src/share/shareLink.test.ts
test('share link mang product_id + ref, không token/PII', async () => {
  jest.spyOn(http, 'get').mockResolvedValue(jsonRes({ code: 'CHI2026' }));
  const link = await buildShareLink(90112);
  expect(link).toBe('https://sandeal.app/p/90112?ref=CHI2026');
  expect(link).not.toMatch(/token|bearer|@/i); // không token, không email
});

test('mobile không tự sinh mã - luôn gọi backend', async () => {
  const getSpy = jest.spyOn(http, 'get').mockResolvedValue(jsonRes({ code: 'CHI2026' }));
  await buildShareLink(90112);
  expect(getSpy).toHaveBeenCalledWith('/v1/referral/code');
});

// apps/mobile/src/deeplink/linkHandler.test.ts
test('parse link lấy product_id + ref', () => {
  const p = parseDeepLink('https://sandeal.app/p/90112?ref=CHI2026');
  expect(p.productId).toBe(90112);
  expect(p.ref).toBe('CHI2026');
});

test('mở link có ref -> lưu pending + route sản phẩm', async () => {
  const nav = { navigate: jest.fn() };
  await handleDeepLink('https://sandeal.app/p/90112?ref=CHI2026', nav as any);
  expect(await readPendingReferral()).toBe('CHI2026');
  expect(nav.navigate).toHaveBeenCalledWith('Product', { productId: 90112 });
});

test('link thiếu product -> màn chủ', async () => {
  const nav = { navigate: jest.fn() };
  await handleDeepLink('https://sandeal.app/p/?ref=CHI2026', nav as any);
  expect(nav.navigate).toHaveBeenCalledWith('Home');
});

// apps/mobile/src/deeplink/pendingReferral.test.ts
test('self-referral bị chặn phía client', async () => {
  await savePendingReferral('CHI2026');
  jest.spyOn(authClient, 'currentUserCodeOrNull').mockResolvedValue('CHI2026'); // chính mình
  const ref = await applyPendingReferralAtSignup();
  expect(ref).toBeNull();
  expect(await readPendingReferral()).toBeNull();
});

test('ref người khác -> giữ để gắn attribution', async () => {
  await savePendingReferral('HUY2026');
  jest.spyOn(authClient, 'currentUserCodeOrNull').mockResolvedValue('CHI2026');
  const ref = await applyPendingReferralAtSignup();
  expect(ref).toBe('HUY2026'); // backend (TASK-BILL-004) ghi attribution
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: shareLink.ts (lấy mã + dựng link) -> ShareSheet.tsx (user-initiated share) -> linkHandler.ts (parse + route) -> pendingReferral.ts (lưu pending + chống self-referral) -> authClient.ts (gắn ref lúc đăng ký) -> cấu hình apple-app-site-association + assetlinks.json -> tests. Universal/app link cần file association host trên domain `sandeal.app`. Backend cần `GET /v1/referral/code` (mặt tiền của TASK-BILL-004) trả mã của user đăng nhập, và luồng đăng ký nhận `ref` để ghi `app_user.referral_code_id`. Test qua Jest với mock http + storage.

---

## §7 - Phụ thuộc

- TASK-MOBILE-001 - scaffold + `httpClient` + deep-link foundation (React Navigation linking).
- TASK-BILL-004 - `referral_code` (mỗi user một mã) + attribution `app_user.referral_code_id` + cấm self-referral + phát sự kiện anti-fraud.
- TASK-MOBILE-002 (liên quan) - màn deal là nơi đặt nút "Chia sẻ".
- TASK-TRUST-004 (downstream của BILL) - anti-fraud nhận sự kiện referral để chống farming.
- Lib: `react-native-share`, React Navigation linking.

---

## §8 - Payload ví dụ

### Lấy mã của user (GET /v1/referral/code)

```json
{ "code": "CHI2026" }
```

### Share link công khai (chỉ product_id + ref)

```
https://sandeal.app/p/90112?ref=CHI2026
```

### Mở app từ link -> ref pending (lưu cục bộ tạm)

```json
{ "pending_referral": "CHI2026", "saved_at": "2026-06-28T09:00:00Z" }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Thẻ xem trước (preview card) đẹp khi share lên mạng xã hội (Open Graph cho web fallback) - tinh chỉnh sau.
- Theo dõi hiệu quả từng kênh chia sẻ (Zalo/Messenger/...) ở mức ẩn danh - thêm khi cần đo viral coefficient.
- Thưởng theo cấp (multi-tier referral) - giữ một cấp ở slice này, mở rộng nếu chiến lược tăng trưởng cần, vẫn qua TASK-BILL-004 + TASK-TRUST-004.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Tự động đăng/gửi link nền | review + test user-initiated | spam, tẩy chay cộng đồng | Chỉ tạo link khi bấm Chia sẻ (DEC-MOBILE-22) |
| Mobile tự sinh referral_code | test luôn gọi backend | mã trùng/giả, bỏ anti-fraud | Lấy mã từ TASK-BILL-004 (DEC-MOBILE-21) |
| Token/PII trong link công khai | test link không token | rò phiên/danh tính | Chỉ product_id + ref ngắn (DEC-MOBILE-25) |
| Self-referral | TestSelf-referral chặn | farming thưởng | Chặn client + backend gác cổng (DEC-MOBILE-24) |
| Mobile tự trả thưởng | review + DEC-MOBILE-24 | gian lận bỏ qua anti-fraud | Thưởng do backend + delay (TASK-TRUST-004) |
| Ghi đè referrer đã có | AC #6 | tranh chấp attribution | Bất biến sau khi gắn (TASK-BILL-004) |
| Deep-link thiếu/sai trường | TestLink thiếu product | app crash/đứng | Mở màn chủ; bỏ qua ref sai |
| Mất attribution giữa click và đăng ký | TestPending lưu/đọc | mất công viral | Lưu ref pending tới lúc đăng ký |

---

## §11 - Ghi chú

- Virality phải user-initiated: người dùng chủ động bấm Chia sẻ; SănDeal không bao giờ tự gửi link (khác spam Honey-style, §5.7).
- Mã referral từ backend (TASK-BILL-004) là nguồn duy nhất - mobile không tự sinh, không tự trả thưởng.
- Universal/app link phục vụ cả người đã cài (mở app) lẫn người mới (web fallback) - đúng mục tiêu viral chạm người chưa có app.
- Attribution gắn một-lần lúc đăng ký, bất biến; ref pending giữ được attribution qua khoảng cài app -> tạo tài khoản.
- Self-referral chặn hai lớp: client (trải nghiệm + giảm tải) và backend (hàng rào thật).
- Link công khai chỉ mang product_id + mã ref ngắn - không bao giờ token/PII của người chia sẻ.
- Chống farming/abuse là việc của TASK-TRUST-004 qua sự kiện từ TASK-BILL-004; mobile chỉ tạo link + gắn attribution.

---

*Hết TASK-MOBILE-003. Status: ready_to_review (awaiting HITL) (mục tiêu audit 10/10).*
