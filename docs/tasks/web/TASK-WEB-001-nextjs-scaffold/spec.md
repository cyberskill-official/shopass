---
id: TASK-WEB-001
title: "Scaffold Next.js (App Router) + tích hợp auth JWT + shell dashboard - khung web app nền cho mọi màn hình SănDeal, gọi BFF qua JWT của TASK-AUTH-002, không tự lưu mật khẩu"
module: WEB
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-AUTH-002, TASK-WEB-002, TASK-WEB-003, TASK-WEB-004, TASK-WEB-005, TASK-INFRA-001]
depends_on: [TASK-AUTH-002]
blocks: [TASK-WEB-002, TASK-WEB-003, TASK-WEB-004, TASK-WEB-005]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.1 (Web App Next.js trong kiến trúc tổng thể)"
  - "docs/... §3.7 (gọi API REST/GraphQL qua gateway), §3.8 (NFR bảo mật no-cleartext, token không rời client server-side)"
source_decisions:
  - "DEC-WEB-01: web app dùng Next.js App Router (Next 14, React Server Components) - layout shell + route group (auth) và (app) tách rõ"
  - "DEC-WEB-02: auth dùng JWT do TASK-AUTH-002 phát; access token ngắn hạn giữ trong bộ nhớ, refresh token trong httpOnly secure SameSite cookie - KHÔNG để token trong localStorage"
  - "DEC-WEB-03: web app KHÔNG tự xác thực mật khẩu hay lưu pwd_hash; mọi credential check ủy quyền cho auth-svc (TASK-AUTH-001/002) qua gateway"
  - "DEC-WEB-04: gọi backend qua một API client tập trung (lib/api.ts) gắn Authorization: Bearer + tự refresh khi 401 một lần; base URL trỏ gateway (TASK-INFRA-001)"
  - "DEC-WEB-05: shell dashboard có route được bảo vệ (middleware) - chưa đăng nhập redirect về /login; không render dữ liệu người dùng phía server khi thiếu phiên hợp lệ"

language: "TypeScript 5.x; Next.js 14 (App Router, React 18); next/server middleware; gọi REST gateway (TASK-INFRA-001)"
service: shopass/web/
new_files:
  - web/package.json
  - web/next.config.mjs
  - web/middleware.ts
  - web/app/layout.tsx
  - web/app/(auth)/login/page.tsx
  - web/app/(app)/dashboard/page.tsx
  - web/app/(app)/layout.tsx
  - web/lib/api.ts
  - web/lib/auth.ts
  - web/components/app-shell.tsx
  - web/test/api-client.test.ts
  - web/test/middleware-guard.test.ts
modified_files:
  - web/.env.example                        # thêm NEXT_PUBLIC_API_BASE_URL trỏ gateway
allowed_tools:
  - file_read: web/**
  - file_write: web/**
  - bash: cd web && npm test && npx tsc --noEmit
disallowed_tools:
  - lưu access/refresh token vào localStorage hoặc sessionStorage (vi phạm DEC-WEB-02, mở cửa XSS đánh cắp token)
  - tự xác thực mật khẩu / lưu pwd_hash ở tầng web (vi phạm DEC-WEB-03 - thuộc auth-svc)
  - render dữ liệu người dùng phía server khi thiếu phiên hợp lệ (vi phạm DEC-WEB-05)
  - hardcode base URL backend thay vì đọc từ env trỏ gateway (vi phạm DEC-WEB-04)

effort_hours: 8
sub_tasks:
  - "1.0h: package.json + next.config.mjs + .env.example (NEXT_PUBLIC_API_BASE_URL) + cấu trúc App Router"
  - "1.0h: app/layout.tsx (root) + components/app-shell.tsx (nav, header, slot nội dung)"
  - "1.5h: lib/auth.ts - lưu access token in-memory, refresh token đọc từ httpOnly cookie qua route handler; helper login/logout/getSession"
  - "1.5h: lib/api.ts - client gắn Bearer, retry refresh một lần khi 401, base URL từ env (gateway)"
  - "1.0h: middleware.ts - guard route group (app); chưa đăng nhập redirect /login"
  - "0.5h: app/(auth)/login/page.tsx - form gọi auth-svc, không tự hash; app/(app)/dashboard/page.tsx shell rỗng"
  - "1.0h: test/api-client.test.ts - gắn Bearer, refresh-on-401 một lần rồi dừng"
  - "0.5h: test/middleware-guard.test.ts - không phiên -> redirect /login; có phiên -> cho qua"

risk_if_skipped: "TASK-WEB-001 là khung mà mọi màn hình web SănDeal mọc lên trên đó - landing SEO (TASK-WEB-002), biểu đồ giá (TASK-WEB-003), quản lý wishlist/alert (TASK-WEB-004), BFF GraphQL (TASK-WEB-005) đều phụ thuộc shell + auth + API client này. Thiếu nó thì không có web app, và theo §5.2 web app là đường sống độc lập với extension nếu sàn chặn extension hay Chrome gỡ extension kiểu Honey - mất nó là mất phương án dự phòng existential. Nếu làm sai bảo mật token (để token trong localStorage) thì một lỗ XSS là chiếm phiên hàng loạt người dùng - mâu thuẫn trực tiếp với cam kết no-cleartext + token-không-rời-client (§3.8) và định vị niềm tin hậu-Honey. Nếu web tự xác thực mật khẩu thay vì ủy quyền auth-svc thì nhân đôi bề mặt tấn công và lệch nguồn sự thật về danh tính."
---

## §1 - Mô tả (BCP-14 normative)

Web app SănDeal **MUST** là một dự án Next.js 14 (App Router) với layout shell dashboard, tích hợp xác thực bằng JWT do TASK-AUTH-002 phát, gọi backend qua một API client tập trung trỏ API Gateway (TASK-INFRA-001), và TUYỆT ĐỐI không lưu token nhạy cảm trong web storage hay tự xác thực mật khẩu. Hợp đồng:

1. Dự án **MUST** dùng Next.js 14 App Router (React Server Components) với hai route group tách biệt: `(auth)` cho trang công khai đăng nhập và `(app)` cho khu vực đã đăng nhập (DEC-WEB-01).
2. `app/layout.tsx` (root layout) **MUST** render khung HTML gốc + provider; `app/(app)/layout.tsx` **MUST** bọc `components/app-shell.tsx` (header, điều hướng, vùng nội dung) cho mọi trang trong khu vực đăng nhập.
3. Access token (JWT ngắn hạn từ TASK-AUTH-002) **MUST** chỉ giữ trong bộ nhớ runtime (biến/React state/closure), KHÔNG ghi vào `localStorage` hay `sessionStorage` (DEC-WEB-02). Refresh token **MUST** nằm trong cookie `httpOnly; Secure; SameSite=Lax`, không đọc được từ JavaScript.
4. Web app **MUST NOT** tự xác thực mật khẩu, KHÔNG lưu `pwd_hash`, KHÔNG tự sinh JWT: trang login chỉ gửi thông tin đăng nhập tới auth-svc (qua gateway) và nhận token (DEC-WEB-03). Mọi kiểm tra credential thuộc TASK-AUTH-001/002.
5. `lib/api.ts` **MUST** là client gọi backend duy nhất: tự gắn header `Authorization: Bearer <access_token>`, đọc base URL từ biến môi trường `NEXT_PUBLIC_API_BASE_URL` trỏ gateway (DEC-WEB-04). KHÔNG hardcode URL backend.
6. Khi một request trả `401`, `lib/api.ts` **MUST** thử làm mới token đúng MỘT lần (gọi luồng refresh của TASK-AUTH-002), rồi phát lại request; nếu vẫn `401`, đăng xuất và chuyển hướng `/login`. KHÔNG lặp refresh vô hạn.
7. `middleware.ts` (Next.js middleware) **MUST** chặn các route trong group `(app)`: nếu không có phiên hợp lệ (thiếu refresh cookie), redirect `307` về `/login` kèm tham số `?next=` để quay lại sau đăng nhập (DEC-WEB-05).
8. Các trang trong `(app)` **MUST NOT** render dữ liệu cá nhân phía server khi thiếu phiên hợp lệ: không gọi API người dùng nếu middleware chưa xác nhận phiên (defense in depth cùng #7).
9. Cấu hình **MUST** đọc base URL gateway từ env (`.env.example` ghi rõ `NEXT_PUBLIC_API_BASE_URL`), không nhúng cứng; build dev và prod dùng cùng client, khác env.
10. `app/(app)/dashboard/page.tsx` **MUST** tồn tại như shell tối thiểu (khung trống, chỗ chứa cho TASK-WEB-003/004) để xác nhận luồng đăng nhập tới khu vực bảo vệ hoạt động đầu-cuối.
11. Dự án **MUST** vượt `npx tsc --noEmit` sạch và `npm test` xanh; cấu hình TypeScript strict (`strict: true`).
12. Web app **SHOULD** đặt nền tảng i18n locale `vi-VN` mặc định (khớp `app_user.locale` ở §3.4) để các task sau gắn chuỗi đa ngôn ngữ khi mở SEA.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao App Router + route group (DEC-WEB-01)?** Tách `(auth)` công khai khỏi `(app)` bảo vệ làm ranh giới bảo mật rõ ràng ở tầng định tuyến: một middleware canh đúng group `(app)`, không phải rải kiểm tra trong từng trang. React Server Components giảm JS gửi xuống client cho phần landing SEO (TASK-WEB-002) - quan trọng cho điểm SEO và tốc độ.

**Vì sao token in-memory + refresh httpOnly (DEC-WEB-02)?** Định vị lõi của SănDeal là niềm tin hậu-Honey và cam kết no-cleartext (§3.8). Token trong `localStorage` đọc được bằng JavaScript - một lỗ XSS là đánh cắp phiên hàng loạt. Access token ngắn hạn giữ trong bộ nhớ (mất khi reload, lấy lại qua refresh) + refresh token trong cookie `httpOnly` (JS không chạm) là mẫu chuẩn giảm bề mặt tấn công token.

**Vì sao web không tự xác thực mật khẩu (DEC-WEB-03)?** Danh tính phải có một nguồn sự thật duy nhất là auth-svc (TASK-AUTH-001/002). Nếu web cũng hash/đối chiếu mật khẩu thì nhân đôi bề mặt tấn công, lệch logic, và vi phạm no-cleartext nếu lỡ lưu. Web chỉ chuyển tiếp thông tin đăng nhập tới auth-svc và tiêu thụ JWT trả về.

**Vì sao một API client tập trung (DEC-WEB-04)?** Gắn Bearer, refresh-on-401, đọc base URL từ env nên nằm một chỗ. Rải logic gọi API khắp các trang dẫn tới quên gắn token, hardcode URL, hoặc refresh không nhất quán. `lib/api.ts` là điểm thắt cho mọi tương tác backend, dễ test và audit.

**Vì sao refresh đúng một lần khi 401 (§1 #6)?** Token có thể hết hạn giữa phiên; một lần refresh rồi phát lại request là đủ. Lặp refresh vô hạn khi backend thực sự từ chối tạo vòng lặp và che lỗi đăng nhập thật. Một lần, rồi đăng xuất sạch, là hành vi đúng.

**Vì sao middleware guard ở tầng định tuyến (DEC-WEB-05)?** Chặn sớm ở middleware (trước khi trang render) rẻ và an toàn hơn để mỗi trang tự kiểm tra. Thiếu phiên thì redirect ngay, không bao giờ chạm API người dùng - tránh rò rỉ dữ liệu phía server cho phiên không hợp lệ.

---

## §3 - Hợp đồng API / DDL

### API client tập trung (lib/api.ts)

```ts
// web/lib/api.ts
const BASE = process.env.NEXT_PUBLIC_API_BASE_URL!; // trỏ gateway (TASK-INFRA-001)

let accessToken: string | null = null; // CHỈ in-memory (DEC-WEB-02)
export function setAccessToken(t: string | null) { accessToken = t; }

export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers);
  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);
  let res = await fetch(`${BASE}${path}`, { ...init, headers });

  if (res.status === 401) {
    const refreshed = await tryRefreshOnce();      // gọi TASK-AUTH-002 refresh
    if (!refreshed) { await logout(); throw new UnauthorizedError(); }
    headers.set("Authorization", `Bearer ${accessToken}`);
    res = await fetch(`${BASE}${path}`, { ...init, headers }); // phát lại MỘT lần
  }
  return res;
}
```

### Refresh + session (lib/auth.ts)

```ts
// web/lib/auth.ts
// Refresh token nằm trong httpOnly cookie; trao đổi qua route handler nội bộ,
// JavaScript KHÔNG đọc giá trị cookie (DEC-WEB-02).
export async function tryRefreshOnce(): Promise<boolean> {
  const res = await fetch("/api/auth/refresh", { method: "POST" }); // route handler đặt/đọc cookie
  if (!res.ok) return false;
  const { accessToken } = await res.json();
  setAccessToken(accessToken); // lưu in-memory
  return true;
}

export async function logout(): Promise<void> {
  setAccessToken(null);
  await fetch("/api/auth/logout", { method: "POST" }); // xóa httpOnly cookie
}
```

### Middleware guard (middleware.ts)

```ts
// web/middleware.ts
import { NextResponse, type NextRequest } from "next/server";

export function middleware(req: NextRequest) {
  const hasSession = req.cookies.has("sd_refresh"); // httpOnly refresh cookie
  if (!hasSession) {
    const url = new URL("/login", req.url);
    url.searchParams.set("next", req.nextUrl.pathname);
    return NextResponse.redirect(url, 307); // DEC-WEB-05
  }
  return NextResponse.next();
}

export const config = { matcher: ["/dashboard/:path*", "/wishlist/:path*", "/alerts/:path*"] };
```

---

## §4 - Acceptance criteria

1. `cd web && npx tsc --noEmit` sạch (TypeScript strict bật) và `npm test` xanh.
2. Cây thư mục có route group `(auth)` và `(app)`; `app/(app)/layout.tsx` bọc `app-shell`.
3. `setAccessToken` lưu token in-memory; grep `web/**`: KHÔNG có `localStorage`/`sessionStorage` chứa token.
4. Trang login gọi auth-svc qua API client; grep: KHÔNG có hashing mật khẩu, KHÔNG có `pwd_hash` ở tầng web.
5. `apiFetch` gắn `Authorization: Bearer` khi có access token và đọc base URL từ `NEXT_PUBLIC_API_BASE_URL`.
6. Khi backend trả `401`, `apiFetch` gọi refresh đúng MỘT lần rồi phát lại; nếu vẫn `401`, đăng xuất và không lặp tiếp.
7. `middleware.ts` redirect `307` về `/login?next=...` khi thiếu cookie `sd_refresh`; cho qua khi có.
8. Truy cập `/dashboard` chưa đăng nhập -> bị redirect `/login`; sau đăng nhập (cookie có) -> render shell dashboard.
9. `.env.example` chứa `NEXT_PUBLIC_API_BASE_URL`; không có URL backend hardcode trong mã.
10. Refresh token chỉ tồn tại trong cookie `httpOnly; Secure; SameSite`; không khóa nào tên token đọc được từ JS phía client.

---

## §5 - Kiểm thử (verification)

```ts
// web/test/api-client.test.ts
import { apiFetch, setAccessToken } from "../lib/api";

test("gắn Authorization: Bearer khi có access token", async () => {
  setAccessToken("abc.def.ghi");
  const spy = mockFetchOk();
  await apiFetch("/v1/wishlists");
  const headers = spy.mock.calls[0][1].headers as Headers;
  expect(headers.get("Authorization")).toBe("Bearer abc.def.ghi");
});

test("401 → refresh đúng một lần rồi phát lại; vẫn 401 → logout, không lặp", async () => {
  setAccessToken("expired");
  const fetchSpy = mockFetchSequence([401, /* refresh */ 200, 401]); // request, refresh, retry
  await expect(apiFetch("/v1/wishlists")).rejects.toThrow(/Unauthorized/);
  // request gốc + 1 refresh + 1 retry = 3 lần fetch, KHÔNG hơn
  expect(fetchSpy).toHaveBeenCalledTimes(3);
});

test("KHÔNG ghi token vào web storage", async () => {
  setAccessToken("xyz");
  expect(globalThis.localStorage?.getItem?.("accessToken")).toBeFalsy();
});
```

```ts
// web/test/middleware-guard.test.ts
import { middleware } from "../middleware";

test("thiếu phiên → redirect 307 /login?next=", () => {
  const req = mockRequest("/dashboard", { cookies: {} });
  const res = middleware(req);
  expect(res.status).toBe(307);
  expect(res.headers.get("location")).toContain("/login?next=%2Fdashboard");
});

test("có phiên → cho qua (next)", () => {
  const req = mockRequest("/dashboard", { cookies: { sd_refresh: "tok" } });
  const res = middleware(req);
  expect(res.headers.get("location")).toBeNull();
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `package.json` + `next.config.mjs` + `.env.example` (cấu hình nền) -> `app/layout.tsx` + `components/app-shell.tsx` (khung hiển thị) -> `lib/auth.ts` (session, refresh httpOnly) -> `lib/api.ts` (client gắn Bearer, refresh-on-401) -> `middleware.ts` (guard group (app)) -> `app/(auth)/login/page.tsx` + `app/(app)/dashboard/page.tsx` (luồng đầu-cuối) -> tests. Route handler `/api/auth/refresh` và `/api/auth/logout` đặt/xóa httpOnly cookie và trao access token in-memory; chúng là cầu nối tới luồng JWT của TASK-AUTH-002 qua gateway (TASK-INFRA-001). Không thêm state management lib ở slice này; React state + closure đủ cho shell.

---

## §7 - Phụ thuộc

- **TASK-AUTH-002** - phát JWT access + refresh; web app tiêu thụ token, không tự sinh (depends_on cứng).
- **TASK-INFRA-001 (gateway)** - mọi gọi backend đi qua gateway; `NEXT_PUBLIC_API_BASE_URL` trỏ tới đây.
- **TASK-WEB-002 (downstream)** - landing SEO mọc trên route group `(auth)`/công khai của scaffold này.
- **TASK-WEB-003 / TASK-WEB-004 (downstream)** - biểu đồ giá và wishlist/alert UI render trong shell `(app)`, gọi backend qua `lib/api.ts`.
- **TASK-WEB-005 (downstream)** - BFF GraphQL là một backend thay thế REST cho client; `lib/api.ts` mở rộng để gọi GraphQL endpoint.
- Lib: `next` 14, `react` 18, `typescript` strict; test qua Jest/Vitest + jsdom.

---

## §8 - Payload ví dụ

### Trang login gọi auth-svc (không tự hash mật khẩu)

```ts
// app/(auth)/login/page.tsx (đoạn submit)
const res = await fetch("/api/auth/login", {            // route handler -> auth-svc qua gateway
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ email, password }),            // gửi nguyên tới auth-svc; web KHÔNG hash
});
if (res.ok) {
  const { accessToken } = await res.json();             // refresh token đã đặt vào httpOnly cookie phía route handler
  setAccessToken(accessToken);                          // lưu in-memory
  router.push(nextParam ?? "/dashboard");
}
```

### Biến môi trường (.env.example)

```
NEXT_PUBLIC_API_BASE_URL=https://api.sandeal.vn
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Chọn thư viện UI/component (giữ tối thiểu ở slice 1) - quyết khi TASK-WEB-003/004 cần biểu đồ và bảng.
- SSR per-request cho dashboard cá nhân hóa vs CSR - đo sau; slice 1 dashboard là shell.
- State management toàn cục (nếu wishlist/alert phình) - thêm khi TASK-WEB-004 cần.
- Đa ngôn ngữ đầy đủ (i18n routing) khi mở SEA - đặt nền locale vi-VN trước (§1 #12).

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Token lưu localStorage | grep test + review | XSS đánh cắp phiên | In-memory access + httpOnly refresh (DEC-WEB-02) |
| Web tự hash/đối chiếu mật khẩu | grep hashing/pwd_hash | nhân đôi bề mặt tấn công | Ủy quyền auth-svc (DEC-WEB-03) |
| Hardcode URL backend | grep http(s):// trong mã | lệch env dev/prod | Đọc NEXT_PUBLIC_API_BASE_URL (DEC-WEB-04) |
| Refresh lặp vô hạn khi 401 | đếm số fetch trong test | vòng lặp, che lỗi auth | Refresh đúng một lần rồi logout (§1 #6) |
| Route bảo vệ render khi chưa đăng nhập | middleware-guard test | rò rỉ dữ liệu phía server | Middleware redirect + không gọi API (DEC-WEB-05) |
| Quên gắn Bearer ở một số call | api-client test | request 401 ngoài ý muốn | Một client tập trung lib/api.ts (DEC-WEB-04) |
| tsc lỗi kiểu | npx tsc --noEmit | build vỡ | strict bật, sửa kiểu trước merge |
| Cookie refresh thiếu Secure/SameSite | review route handler | CSRF/đánh cắp | Đặt httpOnly; Secure; SameSite=Lax |
| Đăng nhập xong không quay lại trang đích | thiếu ?next= | UX kém | middleware gắn next; login đọc next |

---

## §11 - Ghi chú

- Đây là khung nền của toàn bộ web app SănDeal; landing SEO, biểu đồ giá, wishlist/alert, BFF GraphQL đều mọc trên scaffold này.
- Web app là đường sống độc lập extension: nếu sàn chặn extension hay Chrome gỡ extension kiểu Honey (§5.2), web vẫn phục vụ giá trị lõi.
- Token in-memory + refresh httpOnly là mẫu chuẩn no-cleartext; cam kết "token không rời client an toàn" được giữ ở cả phía web, không chỉ extension.
- Web không tự xác thực mật khẩu - một nguồn sự thật danh tính (auth-svc) giảm bề mặt tấn công và tránh lệch logic.
- API client tập trung là điểm thắt cho gắn token + refresh + base URL env, dễ audit và mở rộng sang GraphQL (TASK-WEB-005).

---

*Hết TASK-WEB-001. Status: ready_to_implement (mục tiêu audit 10/10).*
