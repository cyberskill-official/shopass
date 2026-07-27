---
id: TASK-MOBILE-001
title: "Scaffold mobile app (React Native) + auth (JWT của TASK-AUTH-002, lưu token trong secure storage) + đăng ký push FCM/APNs (đăng device token về backend cho TASK-NOTIF-002/005)"
module: MOBILE
priority: SHOULD
status: ready_to_review
verify: T
phase: P3
milestone: P3 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-AUTH-002, TASK-NOTIF-002, TASK-NOTIF-005, TASK-MOBILE-002, TASK-MOBILE-003]
depends_on: [TASK-AUTH-002, TASK-NOTIF-002]
blocks: [TASK-MOBILE-002, TASK-MOBILE-003]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.1 (Mobile App giai đoạn sau, React Native/Flutter; client của API Gateway/BFF qua HTTPS/REST + GraphQL + WSS)"
  - "docs/... §8 (Phase 3: mobile app), §3.6 (FCM/APNs fan-out), §3.8 (no-cleartext, token không rời client an toàn)"
source_decisions:
  - "DEC-MOBILE-01: chọn React Native (TypeScript) cho slice 1 - chia sẻ nhiều khái niệm với web app Next.js (TASK-WEB) + hệ sinh thái push trưởng thành; Flutter là phương án thay thế đã cân nhắc nhưng không chọn ở slice này"
  - "DEC-MOBILE-02: auth dùng JWT access + refresh của TASK-AUTH-002 (KHÔNG phát hành cơ chế auth riêng cho mobile); access token giữ trong bộ nhớ, refresh token lưu trong secure storage của HĐH (Keychain iOS / Keystore Android), KHÔNG lưu cleartext/AsyncStorage thường"
  - "DEC-MOBILE-03: đăng ký push qua FCM cho cả Android và iOS (FCM làm lớp trừu tượng trên APNs) -> backend đã có FCM dispatcher (TASK-NOTIF-002); device token đăng về backend qua POST /v1/devices và gắn user_id qua bảng user_channel_token"
  - "DEC-MOBILE-04: device token được làm tươi mỗi lần khởi động + khi HĐH xoay token; token cũ gỡ liên kết (unregister) khi đăng xuất để không gửi push tới thiết bị đã thoát"
  - "DEC-MOBILE-05: mọi lời gọi backend đi qua API Gateway/BFF (TASK-INFRA-001) bằng HTTPS với access token ở header Authorization; refresh tự động khi 401 do access token hết hạn"
  - "DEC-MOBILE-06: app KHÔNG bao giờ lưu mật khẩu; chỉ giữ token; đăng xuất xóa sạch token khỏi secure storage + bộ nhớ"

language: "React Native 0.74 (TypeScript); push qua @react-native-firebase/messaging; secure storage qua react-native-keychain"
service: shopass/apps/mobile/
new_files:
  - apps/mobile/src/auth/authClient.ts
  - apps/mobile/src/auth/tokenStore.ts
  - apps/mobile/src/auth/authClient.test.ts
  - apps/mobile/src/push/registerDevice.ts
  - apps/mobile/src/push/registerDevice.test.ts
  - apps/mobile/src/api/httpClient.ts
  - apps/mobile/src/api/httpClient.test.ts
  - apps/mobile/src/app/RootNavigator.tsx
modified_files:
  - apps/mobile/package.json                       # thêm dependency firebase messaging + keychain
allowed_tools:
  - file_read: apps/mobile/**
  - file_write: apps/mobile/**
  - bash: cd apps/mobile && npm test
disallowed_tools:
  - lưu refresh token hay mật khẩu trong AsyncStorage thường/cleartext (vi phạm DEC-MOBILE-02/06)
  - phát hành cơ chế auth riêng cho mobile thay vì JWT của TASK-AUTH-002 (vi phạm DEC-MOBILE-02)
  - giữ device token liên kết sau khi đăng xuất (vi phạm DEC-MOBILE-04)

effort_hours: 12
sub_tasks:
  - "1.5h: scaffold RN + TypeScript + cấu trúc thư mục src/{auth,push,api,app}"
  - "2.0h: tokenStore.ts - lưu/đọc/xóa refresh token qua react-native-keychain (Keychain/Keystore); access token chỉ trong bộ nhớ"
  - "2.0h: authClient.ts - login (gọi TASK-AUTH-002), refresh khi 401, logout (xóa token + unregister device)"
  - "2.0h: httpClient.ts - wrapper fetch tới gateway, gắn Authorization, tự refresh + retry một lần khi 401"
  - "2.0h: registerDevice.ts - xin quyền push, lấy FCM token, POST /v1/devices, làm tươi khi token xoay, unregister khi logout"
  - "1.0h: RootNavigator.tsx - điều hướng auth-gated (logged-out -> Login; logged-in -> app shell)"
  - "1.0h: authClient.test.ts + httpClient.test.ts - login/refresh/logout; 401 -> refresh -> retry"
  - "0.5h: registerDevice.test.ts - đăng token; xoay token; unregister khi logout"

risk_if_skipped: "TASK-MOBILE-001 là nền của toàn bộ mobile app (§3.1, §8 Phase 3) - không có scaffold + auth + push thì TASK-MOBILE-002 (theo dõi giá + checkout) và TASK-MOBILE-003 (deep-link virality) không có khung để dựng. Rủi ro bảo mật là điểm dễ sai nhất trên mobile: nếu lưu refresh token trong AsyncStorage thường thay vì secure storage của HĐH thì một thiết bị bị root/jailbreak hoặc một app khác có thể đọc trộm token, vi phạm nguyên tắc no-cleartext + token-an-toàn (§3.8) - đúng cam kết niềm tin hậu-Honey mà SănDeal dựng cả thương hiệu lên. Nếu phát hành cơ chế auth riêng cho mobile thay vì dùng JWT chung thì tạo ra một bề mặt xác thực thứ hai phải bảo trì + kiểm toán song song. Nếu device token không gỡ khi đăng xuất thì push tiếp tục gửi tới thiết bị người dùng đã thoát - rò thông báo cá nhân sang phiên khác."
---

## §1 - Mô tả (BCP-14 normative)

App mobile **MUST** dựng trên React Native (TypeScript), xác thực bằng JWT của TASK-AUTH-002 với refresh token trong secure storage của HĐH, và đăng ký push qua FCM (Android + iOS) với device token đăng về backend. Hợp đồng:

1. **MUST** xác thực bằng JWT access + refresh của TASK-AUTH-002 (DEC-MOBILE-02); **MUST NOT** phát hành cơ chế auth riêng cho mobile.
2. **MUST** lưu refresh token trong secure storage của HĐH (Keychain iOS / Keystore Android qua `react-native-keychain`); access token chỉ giữ trong bộ nhớ tiến trình. **MUST NOT** lưu refresh token hay mật khẩu trong `AsyncStorage` thường hoặc bất kỳ kho cleartext nào.
3. **MUST NOT** bao giờ lưu mật khẩu người dùng trên thiết bị; chỉ giữ token (DEC-MOBILE-06).
4. **MUST** gọi backend qua API Gateway/BFF (TASK-INFRA-001) bằng HTTPS với access token ở header `Authorization: Bearer ...` (DEC-MOBILE-05).
5. **MUST** tự refresh khi nhận `401` do access token hết hạn: dùng refresh token lấy access mới, thử lại request một lần. Nếu refresh thất bại -> đăng xuất (xóa token, về màn Login).
6. **MUST** đăng ký push qua FCM cho cả Android và iOS (FCM trừu tượng hóa APNs) (DEC-MOBILE-03); device token đăng về backend qua `POST /v1/devices` để gắn `user_id` vào `user_channel_token` (nguồn của TASK-NOTIF-002/005).
7. **MUST** xin quyền nhận push của HĐH trước khi đăng token; nếu người dùng từ chối -> không đăng token, app vẫn dùng được (push là kênh phụ trợ, không chặn).
8. **MUST** làm tươi device token mỗi lần khởi động và khi HĐH xoay token (`onTokenRefresh`); token mới đăng về backend, thay token cũ.
9. **MUST** gỡ liên kết device token (unregister) khi đăng xuất (DEC-MOBILE-04): xóa token khỏi secure storage + bộ nhớ, gọi backend gỡ `user_channel_token` để ngừng push tới thiết bị đã thoát.
10. **MUST** cung cấp điều hướng auth-gated: chưa đăng nhập -> màn Login; đã đăng nhập -> app shell (khung cho TASK-MOBILE-002/003).
11. **SHOULD** đo client telemetry tối thiểu (login thành công/thất bại, đăng token thành công/thất bại) ở mức ẩn danh, tôn trọng consent.
12. **MUST** đảm bảo đăng xuất xóa sạch mọi token khỏi secure storage và bộ nhớ (không còn dấu vết phiên).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao React Native, không Flutter ở slice 1 (DEC-MOBILE-01)? Web app SănDeal là Next.js/TypeScript (TASK-WEB). React Native dùng cùng ngôn ngữ + nhiều khái niệm UI, nên đội ngũ chuyển ngữ cảnh ít hơn và chia sẻ được logic (kiểu dữ liệu, client API). Flutter mạnh nhưng kéo theo Dart + một hệ sinh thái thứ hai. Slice 1 chọn React Native để ship nhanh; task ghi rõ Flutter là phương án đã cân nhắc nếu sau này đổi.

Vì sao JWT chung, không auth riêng cho mobile (DEC-MOBILE-02, §1 #1)? Mỗi cơ chế xác thực là một bề mặt phải bảo mật + kiểm toán. Dùng lại JWT access/refresh của TASK-AUTH-002 nghĩa là mobile thừa hưởng nguyên vòng đời token đã được thiết kế (hết hạn, refresh, thu hồi) thay vì phát minh lại. Một nguồn sự thật cho xác thực trên mọi client.

Vì sao refresh token trong secure storage, access chỉ trong bộ nhớ (DEC-MOBILE-02, §1 #2)? Đây là ranh giới bảo mật cốt lõi trên mobile. Refresh token sống lâu nên nếu lộ thì kẻ tấn công duy trì phiên - phải cất trong Keychain/Keystore (vùng được HĐH bảo vệ, gắn với khóa thiết bị). Access token sống ngắn nên giữ trong bộ nhớ tiến trình là đủ; không ghi xuống đĩa giảm bề mặt lộ. AsyncStorage thường là plaintext, đọc được khi thiết bị bị root - tuyệt đối không để token ở đó.

Vì sao FCM cho cả hai nền tảng (DEC-MOBILE-03, §1 #6)? Backend đã có FCM dispatcher (TASK-NOTIF-002). FCM gửi được tới iOS qua APNs ngầm, nên một đường gửi phục vụ cả Android lẫn iOS - bớt một dispatcher phải vận hành ở giai đoạn đầu. TASK-NOTIF-005 (APNs trực tiếp) dành cho khi cần kiểm soát sâu hơn ở iOS.

Vì sao unregister token khi đăng xuất (DEC-MOBILE-04, §1 #9)? Nếu người dùng A đăng xuất rồi người dùng B đăng nhập trên cùng thiết bị (hoặc A bán máy), mà token vẫn liên kết A, thì push của A gửi tới thiết bị giờ thuộc về người khác - rò thông báo cá nhân (giá sản phẩm A theo dõi, đơn hàng A). Gỡ liên kết khi đăng xuất cắt đứt đường rò đó.

Vì sao tự refresh khi 401 (DEC-MOBILE-05, §1 #5)? Access token hết hạn là việc thường xuyên (vòng đời ngắn). Bắt người dùng đăng nhập lại mỗi lần hết hạn là trải nghiệm tệ. Tự refresh trong nền + thử lại một lần làm phiên liền mạch; chỉ khi refresh cũng hỏng (refresh token hết hạn/thu hồi) mới buộc đăng nhập lại.

---

## §3 - Hợp đồng API / mã

### Lưu token an toàn (TypeScript)

```ts
// apps/mobile/src/auth/tokenStore.ts
import * as Keychain from 'react-native-keychain';

const SERVICE = 'sandeal.refresh';

// refresh token CHỈ vào secure storage của HĐH (Keychain/Keystore).
export async function saveRefreshToken(token: string): Promise<void> {
  await Keychain.setGenericPassword('refresh', token, {
    service: SERVICE,
    accessible: Keychain.ACCESSIBLE.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
  });
}

export async function getRefreshToken(): Promise<string | null> {
  const creds = await Keychain.getGenericPassword({ service: SERVICE });
  return creds ? creds.password : null;
}

export async function clearTokens(): Promise<void> {
  await Keychain.resetGenericPassword({ service: SERVICE });
  // access token chỉ ở bộ nhớ - bị xóa khi reset state, không chạm đĩa
}
```

### Đăng ký device push (TypeScript)

```ts
// apps/mobile/src/push/registerDevice.ts
import messaging from '@react-native-firebase/messaging';
import { http } from '../api/httpClient';

// xin quyền -> lấy FCM token -> đăng về backend (gắn user_channel_token).
export async function registerDevice(): Promise<void> {
  const status = await messaging().requestPermission();
  const granted =
    status === messaging.AuthorizationStatus.AUTHORIZED ||
    status === messaging.AuthorizationStatus.PROVISIONAL;
  if (!granted) return; // người dùng từ chối - không đăng, app vẫn chạy

  const token = await messaging().getToken();
  await http.post('/v1/devices', { fcm_token: token, platform: platformName() });

  // HĐH xoay token -> đăng token mới thay token cũ
  messaging().onTokenRefresh(async (next) => {
    await http.post('/v1/devices', { fcm_token: next, platform: platformName() });
  });
}

export async function unregisterDevice(): Promise<void> {
  const token = await messaging().getToken();
  await http.delete('/v1/devices', { fcm_token: token }); // gỡ user_channel_token
  await messaging().deleteToken();
}
```

### Auto-refresh khi 401 (TypeScript)

```ts
// apps/mobile/src/api/httpClient.ts (rút gọn)
async function request(path: string, init: RequestInit, retry = true): Promise<Response> {
  const res = await fetch(GATEWAY + path, withAuth(init, accessToken));
  if (res.status === 401 && retry) {
    const ok = await refreshAccessToken(); // dùng refresh token trong Keychain
    if (!ok) {
      await logout();           // refresh hỏng -> về Login
      return res;
    }
    return request(path, init, false); // thử lại một lần với access mới
  }
  return res;
}
```

---

## §4 - Acceptance criteria

1. App build chạy trên cả Android và iOS (scaffold RN + TypeScript hợp lệ).
2. Login thành công lưu refresh token vào Keychain/Keystore; KHÔNG có refresh token nào trong AsyncStorage thường (kiểm bằng test + review).
3. Access token chỉ tồn tại trong bộ nhớ; không ghi xuống đĩa.
4. Gọi API kèm `Authorization: Bearer <access>`; nhận `401` -> tự refresh -> thử lại một lần -> thành công.
5. Refresh thất bại (refresh token hết hạn) -> đăng xuất, về màn Login.
6. Đăng ký push: xin quyền -> lấy FCM token -> `POST /v1/devices` đăng token + platform.
7. Người dùng từ chối quyền push -> không đăng token, app vẫn dùng được bình thường.
8. HĐH xoay token (`onTokenRefresh`) -> token mới đăng về backend.
9. Đăng xuất -> `clearTokens()` xóa refresh token khỏi Keychain + `unregisterDevice()` gỡ liên kết FCM.
10. Sau đăng xuất, không còn token nào trong secure storage hay bộ nhớ (kiểm bằng test).
11. Không có mật khẩu người dùng nào được ghi xuống thiết bị (review + test).

---

## §5 - Kiểm thử (verification)

```ts
// apps/mobile/src/auth/authClient.test.ts
test('login lưu refresh token vào Keychain, không vào AsyncStorage', async () => {
  await authClient.login('user@example.com', 'pw');
  expect(Keychain.setGenericPassword).toHaveBeenCalled();
  expect(AsyncStorage.setItem).not.toHaveBeenCalledWith(
    expect.stringContaining('refresh'), expect.anything());
});

test('401 -> refresh -> retry một lần thành công', async () => {
  fetchMock.mockResponseOnce('', { status: 401 });
  fetchMock.mockResponseOnce(JSON.stringify({ ok: true }), { status: 200 });
  jest.spyOn(authClient, 'refreshAccessToken').mockResolvedValue(true);
  const res = await http.get('/v1/me');
  expect(res.status).toBe(200);
  expect(authClient.refreshAccessToken).toHaveBeenCalledTimes(1);
});

test('refresh hỏng -> logout, về Login', async () => {
  fetchMock.mockResponse('', { status: 401 });
  jest.spyOn(authClient, 'refreshAccessToken').mockResolvedValue(false);
  const spy = jest.spyOn(authClient, 'logout');
  await http.get('/v1/me');
  expect(spy).toHaveBeenCalled();
});

test('logout xóa sạch token', async () => {
  await authClient.login('user@example.com', 'pw');
  await authClient.logout();
  expect(Keychain.resetGenericPassword).toHaveBeenCalled();
  expect(await getRefreshToken()).toBeNull();
});

// apps/mobile/src/push/registerDevice.test.ts
test('từ chối quyền push -> không đăng token', async () => {
  messaging().requestPermission = jest.fn().mockResolvedValue(
    messaging.AuthorizationStatus.DENIED);
  const postSpy = jest.spyOn(http, 'post');
  await registerDevice();
  expect(postSpy).not.toHaveBeenCalledWith('/v1/devices', expect.anything());
});

test('cấp quyền -> đăng FCM token về backend', async () => {
  messaging().requestPermission = jest.fn().mockResolvedValue(
    messaging.AuthorizationStatus.AUTHORIZED);
  messaging().getToken = jest.fn().mockResolvedValue('fcm_abc');
  const postSpy = jest.spyOn(http, 'post');
  await registerDevice();
  expect(postSpy).toHaveBeenCalledWith('/v1/devices',
    expect.objectContaining({ fcm_token: 'fcm_abc' }));
});

test('logout gỡ liên kết device token', async () => {
  const delSpy = jest.spyOn(http, 'delete');
  await unregisterDevice();
  expect(delSpy).toHaveBeenCalledWith('/v1/devices', expect.objectContaining({ fcm_token: expect.any(String) }));
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: scaffold RN -> tokenStore.ts (Keychain) -> authClient.ts (login/refresh/logout) -> httpClient.ts (auto-refresh) -> registerDevice.ts (FCM) -> RootNavigator.tsx -> tests. Backend cần endpoint `POST /v1/devices` + `DELETE /v1/devices` để gắn/gỡ `user_channel_token` (mở rộng nhẹ từ TASK-NOTIF-002, vốn đã có bảng token). iOS cần cấu hình APNs auth key trong Firebase project để FCM gửi được tới iOS. Test chạy qua Jest với mock của Keychain/messaging/fetch.

---

## §7 - Phụ thuộc

- TASK-AUTH-002 - JWT access + refresh + endpoint login/refresh; mobile dùng lại nguyên cơ chế.
- TASK-NOTIF-002 - FCM dispatcher + bảng `user_channel_token`; mobile đăng device token để dispatcher gửi push.
- TASK-NOTIF-005 (liên quan) - APNs dispatcher trực tiếp; mobile chuẩn bị nền iOS nếu sau này cần.
- TASK-INFRA-001 (gateway) - mọi lời gọi backend qua gateway HTTPS + JWT.
- Lib: `react-native-keychain`, `@react-native-firebase/messaging`, React Navigation.

---

## §8 - Payload ví dụ

### Đăng device token (POST /v1/devices)

```json
{ "fcm_token": "fcm_abc123...", "platform": "ios" }
```

### Gỡ device token khi đăng xuất (DELETE /v1/devices)

```json
{ "fcm_token": "fcm_abc123..." }
```

### Header request có auth

```
GET /v1/me
Authorization: Bearer <access_jwt_ngắn_hạn>
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Đăng nhập sinh trắc học (FaceID/vân tay) mở khóa refresh token - lớp tiện ích thêm sau, vẫn trên nền Keychain.
- Hỗ trợ đăng nhập mạng xã hội trên mobile (Google/Zalo) - tái dùng TASK-AUTH-004 qua deep-link OAuth ở slice sau.
- Đánh giá lại lựa chọn React Native vs Flutter sau slice 1 nếu nhu cầu hiệu năng đồ họa cao.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Refresh token trong AsyncStorage thường | test + review | token lộ khi thiết bị bị root | Bắt buộc Keychain/Keystore (DEC-MOBILE-02) |
| Lưu mật khẩu trên thiết bị | review + test | rò mật khẩu | Chỉ giữ token, không bao giờ mật khẩu (DEC-MOBILE-06) |
| 401 vòng lặp refresh vô hạn | retry=false sau lần đầu | treo app | Chỉ refresh + retry một lần; hỏng -> logout |
| Token không gỡ khi đăng xuất | logout test | push rò sang phiên khác | unregisterDevice() khi logout (DEC-MOBILE-04) |
| Token cũ sau khi HĐH xoay | onTokenRefresh test | push gửi tới token chết | Đăng token mới khi onTokenRefresh |
| Người dùng từ chối quyền push | registerDevice test | không có push | App vẫn chạy; push là kênh phụ trợ |
| Auth riêng cho mobile | review | bề mặt xác thực thứ hai | Dùng lại JWT TASK-AUTH-002 (DEC-MOBILE-02) |
| iOS không nhận push | cấu hình APNs key Firebase | mất kênh iOS | Cấu hình APNs auth key trong Firebase project |

---

## §11 - Ghi chú

- Ranh giới bảo mật cốt lõi mobile: refresh token trong secure storage HĐH, access token chỉ trong bộ nhớ, không bao giờ mật khẩu trên đĩa.
- Một nguồn sự thật cho xác thực: JWT của TASK-AUTH-002 dùng chung web + mobile, không phát minh lại.
- Một đường push (FCM) phục vụ cả Android + iOS ở giai đoạn đầu; TASK-NOTIF-005 (APNs) cho khi cần kiểm soát sâu.
- Gỡ device token khi đăng xuất là biện pháp chống rò thông báo cá nhân sang phiên/người dùng khác trên cùng thiết bị.
- Tự refresh khi 401 giữ phiên liền mạch; chỉ buộc đăng nhập lại khi refresh token cũng hỏng.
- task này là khung; theo dõi giá + checkout (TASK-MOBILE-002) và deep-link virality (TASK-MOBILE-003) dựng lên trên.

---

*Hết TASK-MOBILE-001. Status: ready_to_review (awaiting HITL) (mục tiêu audit 10/10).*
