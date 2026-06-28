---
id: FR-INFRA-001
title: "API Gateway / BFF - định tuyến REST+GraphQL+WSS, rate-limit per-user/IP, verify JWT (uỷ quyền AUTH), WAF rules, request-id propagation"
module: INFRA
priority: MUST
status: ready_to_implement
verify: T
phase: P0
milestone: P0 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-INFRA-002, FR-INFRA-003, FR-INFRA-004, FR-AUTH-002, FR-WEB-005, FR-EXT-005]
depends_on: []
blocks: [FR-AUTH-002, FR-B2B-004, FR-INFRA-004, FR-WEB-005]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.1 (kiến trúc tổng thể, API Gateway/BFF)"
  - "docs/... §3.7 (hợp đồng API REST + GraphQL + WSS), §3.8 (NFR hiệu năng/bảo mật)"
source_decisions:
  - "DEC-INFRA-01: một edge duy nhất (API Gateway/BFF) cho REST + GraphQL + WSS; mọi client (extension, web, mobile) đi qua đây"
  - "DEC-INFRA-02: gateway verify JWT bằng JWKS lấy từ AUTH; KHÔNG tự ký token (uỷ quyền phát hành cho FR-AUTH-002)"
  - "DEC-INFRA-03: rate-limit hai trục - per-user (sub trong JWT) và per-IP (ẩn danh) - token-bucket trong Redis"
  - "DEC-INFRA-04: mỗi request gắn một X-Request-Id; propagate sang downstream service và log; là gốc của trace (FR-INFRA-004)"
  - "DEC-INFRA-05: WAF rule lớp edge (chặn path traversal, SQLi pattern thô, body-size cap, method allowlist) trước khi tới service"

language: "Go 1.22 (gateway-svc); Envoy/Kong tuỳ chọn hạ tầng nhưng hợp đồng test ở tầng Go middleware"
service: shopass/services/gateway/
new_files:
  - services/gateway/internal/gw/router.go
  - services/gateway/internal/gw/jwt.go
  - services/gateway/internal/gw/ratelimit.go
  - services/gateway/internal/gw/waf.go
  - services/gateway/internal/gw/requestid.go
  - services/gateway/internal/gw/router_test.go
  - services/gateway/internal/gw/ratelimit_test.go
  - services/gateway/internal/gw/jwt_test.go
modified_files:
  - services/gateway/cmd/gateway/main.go            # wire middleware chain
allowed_tools:
  - file_read: services/gateway/**
  - file_write: services/gateway/**
  - bash: cd services/gateway && go test ./...
disallowed_tools:
  - tự phát hành/tự ký JWT trong gateway (vi phạm DEC-INFRA-02; phát hành thuộc AUTH)
  - lưu secret JWKS/khoá trong code hay env (phải lấy runtime qua FR-INFRA-003)
  - bỏ qua rate-limit cho path không xác thực (mở đường brute-force/DDoS)

effort_hours: 8
sub_tasks:
  - "1.0h: router.go - định tuyến REST `/v1/*`, GraphQL `/graphql`, WSS `/ws` tới upstream theo prefix"
  - "1.5h: jwt.go - verify chữ ký bằng JWKS cache, kiểm exp/iss/aud, gắn claims vào context"
  - "1.5h: ratelimit.go - token-bucket Redis per-user + per-IP, trả 429 + Retry-After"
  - "1.0h: waf.go - method allowlist, body-size cap, chặn pattern path traversal/SQLi thô"
  - "0.5h: requestid.go - sinh/nhận X-Request-Id, propagate header downstream"
  - "1.0h: jwt_test.go - token hợp lệ pass, hết hạn/sai chữ ký/sai aud bị từ chối 401"
  - "1.0h: ratelimit_test.go - vượt ngưỡng -> 429; bucket per-user tách per-IP"
  - "0.5h: router_test.go - định tuyến đúng upstream theo prefix; WSS upgrade pass-through"

risk_if_skipped: "Không có edge thống nhất thì mỗi service tự lo auth, rate-limit và WAF - trùng lặp, dễ hở. Không verify JWT ở edge -> request giả mạo lọt vào service nội bộ. Không rate-limit -> brute-force đăng nhập và DDoS đốt tài nguyên/đốt chi phí scraping gián tiếp. Không propagate request-id -> không lần được vết sự cố xuyên service (vỡ nền tảng observability FR-INFRA-004). Đây là cổng vào của toàn hệ thống; mọi module P1+ phụ thuộc nó."
---

## §1 - Mô tả (BCP-14 normative)

Service INFRA **MUST** cung cấp một API Gateway/BFF làm điểm vào duy nhất cho REST, GraphQL và WebSocket, thực thi xác thực JWT, rate-limit, WAF lớp edge, và truyền request-id xuyên downstream. Hợp đồng:

1. Gateway **MUST** định tuyến theo prefix: `/v1/*` -> REST upstream tương ứng (auth, price, track, ...); `/graphql` -> BFF GraphQL (FR-WEB-005); `/ws` -> kênh WebSocket (FR-EXT-005). Mọi client (extension, web, mobile) đi qua một host edge (DEC-INFRA-01).
2. Gateway **MUST** verify JWT trên mọi route được bảo vệ: lấy public key qua JWKS endpoint của AUTH, kiểm chữ ký, `exp`, `iss`, `aud`. KHÔNG tự ký hay phát hành token (DEC-INFRA-02) - phát hành thuộc FR-AUTH-002.
3. Gateway **MUST** từ chối với `401 Unauthorized` khi token thiếu/hết hạn/sai chữ ký/sai `aud` trên route bảo vệ; route công khai (ví dụ `/v1/health`, landing API đọc) được khai báo allowlist rõ ràng.
4. Gateway **MUST** gắn claims đã verify (`user_id`, `locale`, `tier`) vào request context và truyền xuống downstream qua header nội bộ đã ký/tin cậy (ví dụ `X-User-Id`), để service nội bộ không phải tự verify lại.
5. Gateway **MUST** rate-limit hai trục bằng token-bucket trong Redis (DEC-INFRA-03): per-user (khoá theo `sub`) cho request đã xác thực; per-IP cho request ẩn danh. Vượt ngưỡng -> `429 Too Many Requests` kèm header `Retry-After`.
6. Gateway **MUST** đặt ngưỡng rate-limit cấu hình được per-route-class (ví dụ auth-login chặt hơn read-API); mặc định an toàn áp cho mọi route chưa khai báo.
7. Gateway **MUST** sinh `X-Request-Id` (UUIDv4) nếu client chưa gửi, hoặc dùng lại giá trị client gửi; **MUST** propagate header này sang mọi downstream call và đưa vào structured log (DEC-INFRA-04). Đây là gốc tương quan trace của FR-INFRA-004.
8. Gateway **MUST** áp WAF lớp edge (DEC-INFRA-05): method allowlist (`GET/POST/PUT/PATCH/DELETE/OPTIONS`); body-size cap cấu hình được (mặc định 1 MiB cho REST); chặn pattern path traversal (`../`) và SQLi thô trong query/path. Vi phạm -> `400` hoặc `413`.
9. Gateway **MUST** xử lý WebSocket upgrade ở `/ws`: verify JWT trong handshake (query param hoặc subprotocol), từ chối handshake không xác thực; sau khi mở, pass-through frame tới upstream realtime.
10. Gateway **MUST** lấy JWKS và mọi secret runtime qua FR-INFRA-003 (Vault/Secrets Manager), KHÔNG nhúng trong code/env.
11. Gateway **SHOULD** phát OTel metric: `gw_requests_total{route_class, status}` (counter), `gw_request_duration_ms` (histogram), `gw_ratelimited_total{axis=user|ip}` (counter), `gw_jwt_rejected_total{reason}` (counter).
12. Gateway **MUST** cache JWKS với TTL và tự refresh khi gặp `kid` chưa biết (hỗ trợ xoay khoá của AUTH mà không downtime).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao một edge duy nhất (DEC-INFRA-01)? Extension MV3, web Next.js và mobile sau này đều cần cùng một bộ quy tắc auth, rate-limit và WAF. Nếu mỗi service tự lo, ta lặp logic ở nhiều nơi và mỗi nơi là một chỗ hở. Một gateway tập trung các mối quan tâm xuyên suốt, để service nội bộ chỉ lo nghiệp vụ.

Vì sao gateway verify nhưng KHÔNG phát hành JWT (DEC-INFRA-02)? Phát hành token là việc của AUTH (FR-AUTH-002): nó giữ khoá ký, biết vòng đời phiên, biết refresh. Gateway chỉ cần public key (JWKS) để verify. Tách hai vai trò này tránh nhân đôi khoá ký và giữ ranh giới tin cậy gọn.

Vì sao rate-limit hai trục (DEC-INFRA-03)? Request đã xác thực nên giới hạn theo người dùng (`sub`) để một tài khoản không lạm dụng. Request ẩn danh (đăng ký, login, landing API) chưa có `sub` nên phải giới hạn theo IP để chặn brute-force và bot. Hai trục bù cho nhau: per-user chặn lạm dụng nội bộ, per-IP chặn tấn công ngoài.

Vì sao Redis token-bucket? Token-bucket cho phép burst ngắn nhưng giới hạn trung bình, hợp với pattern dùng thật (người dùng bấm vài request liền rồi nghỉ). Redis chia sẻ trạng thái bucket giữa nhiều instance gateway, nên rate-limit đúng kể cả khi scale ngang.

Vì sao request-id propagation (DEC-INFRA-04)? Một thao tác người dùng (ví dụ "theo dõi sản phẩm") đi qua gateway -> track-svc -> price-svc -> scraping. Khi lỗi, không có một id chung thì không lần được chuỗi. `X-Request-Id` là sợi chỉ xuyên suốt; FR-INFRA-004 nối nó vào trace OTel.

Vì sao WAF lớp edge (DEC-INFRA-05)? Chặn lớp thô (path traversal, SQLi pattern, body khổng lồ) ở rìa rẻ hơn và an toàn hơn là để mọi service tự phòng. Đây không thay được input validation trong service, nhưng cắt phần lớn rác và tấn công tự động trước khi tốn tài nguyên downstream.

Vì sao cache JWKS có refresh theo `kid` (§1 #12)? AUTH sẽ xoay khoá ký định kỳ. Nếu gateway cache cứng, token ký bằng khoá mới sẽ bị từ chối oan. Bắt gặp `kid` lạ thì refresh JWKS giúp xoay khoá mượt, không downtime.

---

## §3 - Hợp đồng API / DDL

### Middleware chain (Go)

```go
// services/gateway/internal/gw/router.go
// Thứ tự middleware: requestID → waf → rateLimit → jwtVerify → routeUpstream.
func NewHandler(deps Deps) http.Handler {
    mux := http.NewServeMux()
    mux.Handle("/v1/", upstreamREST(deps))
    mux.Handle("/graphql", upstreamGraphQL(deps))
    mux.Handle("/ws", wsUpgrade(deps)) // verify JWT trong handshake
    return chain(
        requestID(),            // §1 #7
        waf(deps.WAFConfig),    // §1 #8
        rateLimit(deps.Redis),  // §1 #5,#6
        jwtVerify(deps.JWKS),   // §1 #2,#3,#4 (bỏ qua trên allowlist công khai)
    )(mux)
}
```

### JWT verify (§1 #2, #3, #4, #12)

```go
// services/gateway/internal/gw/jwt.go
type Claims struct {
    UserID int64  `json:"user_id"`
    Locale string `json:"locale"`
    Tier   string `json:"tier"`
    jwt.RegisteredClaims
}

func jwtVerify(jwks *JWKSCache) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if isPublic(r.URL.Path) { next.ServeHTTP(w, r); return }
            raw := bearer(r)
            claims, err := jwks.Verify(r.Context(), raw) // kiểm sig/exp/iss/aud, refresh nếu kid lạ
            if err != nil {
                metrics.JWTRejected(reasonOf(err))
                writeJSON(w, 401, errBody("unauthorized"))
                return
            }
            r = r.WithContext(withClaims(r.Context(), claims))
            r.Header.Set("X-User-Id", strconv.FormatInt(claims.UserID, 10)) // §1 #4
            next.ServeHTTP(w, r)
        })
    }
}
```

### Rate-limit token-bucket (§1 #5, #6)

```go
// services/gateway/internal/gw/ratelimit.go
// Khoá bucket: "rl:user:<sub>" nếu đã xác thực, ngược lại "rl:ip:<ip>".
// Dùng Redis script INCR + EXPIRE (token-bucket xấp xỉ theo cửa sổ trượt).
func rateLimit(rdb RedisClient) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key, limit := bucketKeyAndLimit(r) // per-route-class
            ok, retryAfter, err := rdb.AllowN(r.Context(), key, limit)
            if err == nil && !ok {
                w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
                metrics.RateLimited(axisOf(key))
                writeJSON(w, 429, errBody("rate_limited"))
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## §4 - Acceptance criteria

1. Request `GET /v1/health` (public allowlist) -> 200, không yêu cầu JWT.
2. Request route bảo vệ không kèm Bearer -> 401.
3. Request route bảo vệ với JWT hết hạn -> 401, metric `gw_jwt_rejected_total{reason="expired"}` tăng.
4. Request với JWT sai chữ ký -> 401.
5. Request với JWT sai `aud` -> 401.
6. Request với JWT hợp lệ -> pass; downstream nhận header `X-User-Id` đúng giá trị `sub`.
7. Bắn vượt ngưỡng per-IP trên `/v1/auth/login` -> 429 kèm `Retry-After`.
8. Bucket per-user của user A không làm cạn bucket per-user của user B (cô lập theo `sub`).
9. Request POST body > cap (mặc định 1 MiB) -> 413.
10. Request path chứa `../` -> 400 (WAF chặn).
11. Method không trong allowlist (ví dụ `TRACE`) -> 405/400.
12. Request không gửi `X-Request-Id` -> gateway sinh UUIDv4 và propagate; downstream thấy cùng id trong log.
13. WSS handshake ở `/ws` không kèm JWT hợp lệ -> handshake bị từ chối; kèm JWT hợp lệ -> upgrade thành công.
14. JWKS xoay `kid` mới -> gateway refresh và verify token mới thành công không cần restart.

---

## §5 - Kiểm thử (verification)

```go
// services/gateway/internal/gw/jwt_test.go
func TestJWT_Expired_401(t *testing.T) {
    h := newTestHandler(t)
    tok := signToken(t, claimsExpired())
    rr := do(h, "GET", "/v1/track", tok)
    require.Equal(t, 401, rr.Code)
}

func TestJWT_BadAudience_401(t *testing.T) {
    h := newTestHandler(t)
    tok := signToken(t, claimsAud("other-service"))
    rr := do(h, "GET", "/v1/track", tok)
    require.Equal(t, 401, rr.Code)
}

func TestJWT_Valid_PropagatesUserID(t *testing.T) {
    h, upstream := newTestHandlerWithUpstream(t)
    tok := signToken(t, claimsValid(90112))
    rr := do(h, "GET", "/v1/track", tok)
    require.Equal(t, 200, rr.Code)
    require.Equal(t, "90112", upstream.LastHeader("X-User-Id"))
}

// services/gateway/internal/gw/ratelimit_test.go
func TestRateLimit_PerIP_429(t *testing.T) {
    h := newTestHandler(t)
    for i := 0; i < loginLimit; i++ {
        require.Equal(t, 200, doIP(h, "POST", "/v1/auth/login", "1.2.3.4").Code)
    }
    over := doIP(h, "POST", "/v1/auth/login", "1.2.3.4")
    require.Equal(t, 429, over.Code)
    require.NotEmpty(t, over.Header().Get("Retry-After"))
}

func TestRateLimit_PerUser_Isolated(t *testing.T) {
    h := newTestHandler(t)
    tokA, tokB := signUser(t, 1), signUser(t, 2)
    for i := 0; i < userLimit; i++ { do(h, "GET", "/v1/track", tokA) }
    require.Equal(t, 200, do(h, "GET", "/v1/track", tokB).Code) // B không bị A làm cạn
}

func TestWAF_PathTraversal_400(t *testing.T) {
    h := newTestHandler(t)
    require.Equal(t, 400, do(h, "GET", "/v1/../etc/passwd", "").Code)
}

func TestRequestID_Generated(t *testing.T) {
    h, upstream := newTestHandlerWithUpstream(t)
    do(h, "GET", "/v1/health", "")
    require.NotEmpty(t, upstream.LastHeader("X-Request-Id"))
}
```

---

## §6 - Khung triển khai

Xem §3. Middleware chain theo thứ tự requestID -> waf -> rateLimit -> jwtVerify -> routeUpstream (requestID đầu tiên để mọi từ chối sau đó vẫn có id để log). JWKS cache với TTL ngắn (ví dụ 10 phút) cộng refresh khi `kid` lạ. Rate-limit dùng Redis dùng chung giữa các instance gateway (không in-memory, vì sẽ sai khi scale ngang). Nếu hạ tầng dùng Envoy/Kong, hợp đồng test vẫn ở tầng Go middleware để CI kiểm được không phụ thuộc một proxy cụ thể.

---

## §7 - Phụ thuộc

- FR-INFRA-003 (đồng hàng) - cung cấp JWKS URL/khoá runtime qua Vault; gateway không nhúng secret.
- FR-AUTH-002 (downstream) - phát hành JWT + JWKS endpoint mà gateway verify.
- FR-INFRA-004 (downstream) - nối `X-Request-Id` thành trace OTel; đọc metric gateway.
- FR-WEB-005 / FR-EXT-005 (downstream) - GraphQL BFF và WSS đi qua route `/graphql`, `/ws`.
- Hạ tầng: Redis (rate-limit state); tuỳ chọn Envoy/Kong làm edge proxy.

---

## §8 - Payload ví dụ

### Request đã xác thực

```http
GET /v1/products/90112/price-history?range=90d HTTP/1.1
Host: api.sandeal.vn
Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0yMDI2LTA2In0...
X-Request-Id: 6b1d...c004
```

### Phản hồi bị rate-limit

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 12
Content-Type: application/json

{"error":"rate_limited","request_id":"6b1d...c004"}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- mTLS giữa gateway và service nội bộ (so với header `X-User-Id` đã ký) - siết zero-trust ở slice hạ tầng sau.
- Adaptive rate-limit theo điểm rủi ro (chặt hơn khi nghi bot) - gắn vào FR-TRUST-004 giai đoạn sau.
- Circuit-breaker per-upstream - thêm khi có nhiều service và cần cách ly lỗi.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| JWKS endpoint của AUTH down | lỗi fetch + metric | không verify được token mới | Phục vụ từ cache TTL; cảnh báo; AUTH HA |
| `kid` lạ (đã xoay khoá) | verify fail lần đầu | 401 oan nếu không refresh | Refresh JWKS khi gặp kid lạ (§1 #12) |
| Redis down | lỗi AllowN | rate-limit hở | Fail-open có ngưỡng + cảnh báo; hoặc fail-closed cho route nhạy cảm |
| Token hợp lệ nhưng user bị khoá | gateway không biết status | request lọt vào service | Service kiểm status (FR-AUTH-005); token TTL ngắn |
| Body khổng lồ (DoS) | WAF body-size | 413 trước khi tốn downstream | Cap cấu hình; streaming reject sớm |
| Path traversal / SQLi thô | WAF pattern | 400 ở edge | Vẫn validate trong service (defense in depth) |
| Thiếu X-Request-Id từ client | requestID middleware | mất tương quan | Gateway tự sinh UUIDv4 |
| Scale ngang dùng in-memory bucket | rate sai (mỗi node riêng) | giới hạn lỏng | Bắt buộc Redis dùng chung (§6) |
| WSS giữ kết nối lạm dụng | đếm kết nối/giây | cạn tài nguyên | Giới hạn handshake theo IP/user; idle timeout |

---

## §11 - Ghi chú

- Gateway là cổng vào duy nhất: mọi client và mọi module P1+ đi qua đây, nên nó là điểm áp chính sách auth/rate-limit/WAF tập trung.
- Ranh giới tin cậy gọn: AUTH phát hành (giữ khoá ký), gateway verify (chỉ public key qua JWKS), service nội bộ tin header đã ký từ gateway.
- Rate-limit hai trục là tuyến phòng thủ kép: per-user chặn lạm dụng tài khoản, per-IP chặn tấn công ẩn danh.
- `X-Request-Id` là nền của observability - không có nó thì FR-INFRA-004 không tương quan được trace xuyên service.
- WAF edge cắt rác/tấn công tự động trước downstream, nhưng không thay input validation trong từng service.

---

*Hết FR-INFRA-001. Status: ready_to_implement (mục tiêu audit 10/10).*
