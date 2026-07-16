---
id: TASK-INFRA-004
title: "Observability spine - Prometheus + Grafana + OTel tracing xuyên service + structured JSON logs + trace-id correlation"
module: INFRA
priority: MUST
status: done
verify: T
phase: P0
milestone: P0 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-INFRA-001, TASK-INFRA-003, TASK-COMPLY-004, TASK-PRICE-002, TASK-SCRAPE-001, TASK-NOTIF-003]
depends_on: [TASK-INFRA-001]
blocks: [TASK-COMPLY-004]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.8 (observability: Prometheus + Grafana + structured logs + tracing OpenTelemetry)"
  - "docs/... §3.1 (kiến trúc nhiều service), §5.5 (breach 72h cần điều tra nhanh)"
source_decisions:
  - "DEC-INFRA-16: tracing dùng OpenTelemetry (OTel) xuyên mọi service; mỗi span mang trace_id chung bắt nguồn từ X-Request-Id của gateway"
  - "DEC-INFRA-17: metric phơi bày theo chuẩn Prometheus (pull /metrics); Grafana là lớp hiển thị + cảnh báo"
  - "DEC-INFRA-18: log là structured JSON một-dòng-một-sự-kiện, BẮT BUỘC chứa trace_id + request_id để tương quan với trace"
  - "DEC-INFRA-19: dùng OTel propagation (W3C traceparent) để trace nối liền qua HTTP giữa các service"
  - "DEC-INFRA-20: log KHÔNG chứa dữ liệu cá nhân thô (PDPL) - mask email/phone/token; chỉ ghi id ẩn danh"

language: "Go 1.22 (shared obs package); OpenTelemetry SDK; Prometheus client; Grafana"
service: shopass/obs/
new_files:
  - obs/tracing.go
  - obs/metrics.go
  - obs/logging.go
  - obs/middleware.go
  - obs/redact.go
  - obs/tracing_test.go
  - obs/logging_test.go
  - obs/redact_test.go
  - deploy/grafana/dashboards/overview.json
modified_files:
  - services/gateway/cmd/gateway/main.go            # khởi tạo obs + middleware đầu chain
allowed_tools:
  - file_read: obs/**
  - file_read: deploy/grafana/**
  - file_write: obs/**
  - file_write: deploy/grafana/**
  - bash: cd obs && go test ./...
disallowed_tools:
  - log dữ liệu cá nhân thô (email/phone/token) (vi phạm DEC-INFRA-20, PDPL)
  - tạo trace_id mới giữa chừng làm đứt chuỗi xuyên service (vi phạm DEC-INFRA-16/19)
  - phơi bày /metrics ra public không bảo vệ (rò rỉ tín hiệu nội bộ)

effort_hours: 8
sub_tasks:
  - "1.5h: tracing.go - khởi tạo OTel TracerProvider, exporter OTLP, W3C traceparent propagation"
  - "1.0h: metrics.go - registry Prometheus + helper counter/histogram + handler /metrics"
  - "1.5h: logging.go - slog JSON handler tự gắn trace_id/request_id từ context"
  - "1.0h: middleware.go - HTTP middleware tạo span gốc từ X-Request-Id, inject vào downstream"
  - "1.0h: redact.go - mask email/phone/token trong field log (PDPL)"
  - "1.0h: tracing_test.go - trace nối liền qua 2 service giả (cùng trace_id); traceparent propagate"
  - "0.5h: logging_test.go - mọi log line có trace_id + request_id; JSON hợp lệ"
  - "0.5h: redact_test.go - email/phone/token không xuất hiện thô trong log"
  - "1.0h: overview.json - Grafana dashboard p95 latency, error rate, request volume per-service"

risk_if_skipped: "Hệ thống nhiều service (gateway, auth, price, track, scrape, notif). Khi sự cố, không có trace xuyên service thì không lần được một request đi đâu, vỡ ở đâu - điều tra mò mẫm hàng giờ. Không có metric/dashboard thì không thấy p95 vỡ NFR hay error spike cho tới khi user than. PDPL yêu cầu báo cáo vi phạm trong 72 giờ (§5.5) - không quan sát được thì không phát hiện kịp, không điều tra kịp. Đây là bề mặt điều tra sự cố nền tảng mà mọi service và TASK-COMPLY-004 dựa vào."
---

## §1 - Mô tả (BCP-14 normative)

Service INFRA **MUST** dựng xương sống quan sát: tracing OTel xuyên service, metric Prometheus, log JSON có cấu trúc, tất cả tương quan qua một trace_id chung bắt nguồn từ gateway. Hợp đồng:

1. Hệ thống **MUST** dùng OpenTelemetry cho tracing xuyên mọi service (DEC-INFRA-16). Mỗi span mang `trace_id` chung; trace_id bắt nguồn từ `X-Request-Id` của gateway (TASK-INFRA-001 #7) để một thao tác người dùng có một trace liền mạch.
2. Trace **MUST** nối liền qua HTTP giữa các service bằng W3C `traceparent` propagation (DEC-INFRA-19): service downstream tiếp tục span của service upstream, KHÔNG tạo trace_id mới giữa chừng.
3. Mỗi service **MUST** phơi bày endpoint `/metrics` theo chuẩn Prometheus (pull model) (DEC-INFRA-17). Endpoint `/metrics` **MUST** được bảo vệ (mạng nội bộ hoặc auth), KHÔNG public trần.
4. Helper metric **MUST** cung cấp counter và histogram tiện dụng để mọi service phát metric đồng nhất (tên, nhãn theo quy ước): ví dụ `http_requests_total{service, route, status}`, `http_request_duration_ms{service, route}`.
5. Log **MUST** là structured JSON, một dòng một sự kiện (DEC-INFRA-18). Mỗi log line **MUST** chứa `trace_id` và `request_id` (lấy từ context) để join log với trace.
6. Log **MUST** KHÔNG chứa dữ liệu cá nhân thô (DEC-INFRA-20, PDPL): email, phone, token, cookie phải được mask hoặc thay bằng id ẩn danh trước khi ghi.
7. HTTP middleware obs **MUST** đứng đầu chain mỗi service: tạo (hoặc tiếp tục) span gốc từ header đến, gắn `trace_id`/`request_id` vào context, đóng span với status code khi xong.
8. Khi gọi downstream, service **MUST** inject `traceparent` vào outgoing request để chuỗi trace tiếp tục.
9. Grafana **MUST** có ít nhất một dashboard tổng quan: p95 latency per-service, error rate, request volume - đọc từ metric Prometheus; dashboard versioned trong repo (`deploy/grafana/dashboards/`).
10. Lớp obs **SHOULD** hỗ trợ sampling cấu hình được cho trace (ví dụ luôn lấy mẫu trace có lỗi; sample tỉ lệ với trace thành công) để cân bằng chi phí lưu trace với độ phủ.
11. Lớp obs **MUST** khởi tạo idempotent và shutdown sạch (flush span/metric khi service dừng) để không mất tín hiệu cuối.
12. Tài liệu/đối chiếu NFR: dashboard **MUST** hiển thị các ngưỡng của NFR-INFRA-001 (p95 < 300ms đọc cache, biểu đồ < 500ms) để vi phạm thấy được ngay.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao trace xuyên service từ một trace_id chung (DEC-INFRA-16)? Một thao tác "theo dõi sản phẩm" đi gateway -> track-svc -> price-svc -> scraping. Khi nó chậm hay lỗi, câu hỏi đầu tiên là "chậm/vỡ ở chặng nào". Không có trace_id chung, mỗi service là một hộp đen riêng. Bắt nguồn trace_id từ `X-Request-Id` của gateway nối client tới tận đáy backend bằng một sợi chỉ.

Vì sao W3C traceparent (DEC-INFRA-19)? Đây là chuẩn propagation để span của service B nối tiếp span của service A. Nếu mỗi service tự sinh trace_id, chuỗi đứt và ta được nhiều trace rời rạc thay vì một trace đầu-cuối. Traceparent giữ quan hệ cha-con span qua ranh giới HTTP.

Vì sao Prometheus pull + Grafana (DEC-INFRA-17)? Pull model (`/metrics`) là chuẩn vận hành phổ biến, dễ scrape, dễ vận hành HA. Grafana tách hiển thị/cảnh báo khỏi thu thập. Tách lớp này cho ta đổi backend lưu trữ metric mà không sửa cách service phát.

Vì sao log JSON có trace_id (DEC-INFRA-18)? Metric cho "cái gì sai" (error rate tăng), trace cho "ở chặng nào", log cho "chi tiết gì". Ba cái chỉ mạnh khi join được. Bắt buộc `trace_id`/`request_id` trong mọi log line là cách join: thấy một trace lỗi trên Grafana -> nhảy thẳng tới log của đúng trace đó.

Vì sao log không chứa dữ liệu cá nhân thô (DEC-INFRA-20)? PDPL coi email/phone là dữ liệu cá nhân; log thường bị lưu lâu, sao chép, chia sẻ rộng - nơi rò rỉ âm thầm. Mask trước khi ghi giữ log hữu ích để debug (vẫn có id ẩn danh để tương quan) mà không biến log thành kho dữ liệu cá nhân vi phạm luật.

Vì sao đối chiếu ngưỡng NFR trên dashboard (§1 #12)? NFR-INFRA-001 đặt p95 < 300ms / biểu đồ < 500ms. Một con số NFR chỉ có nghĩa khi đo được và thấy khi vỡ. Đưa ngưỡng vào dashboard biến NFR trừu tượng thành đường kẻ đỏ trực quan, phát hiện hồi quy sớm.

Vì sao sampling cấu hình được (§1 #10)? Lưu mọi trace ở quy mô lớn tốn kém. Luôn lấy trace có lỗi (giá trị điều tra cao) và sample tỉ lệ trace thành công cân bằng chi phí với độ phủ - vẫn thấy bức tranh mà không lưu tất cả.

---

## §3 - Hợp đồng API / DDL

### Tracing init + propagation (Go)

```go
// obs/tracing.go
func InitTracer(ctx context.Context, serviceName string) (shutdown func(context.Context) error, err error) {
    exp, err := otlptracegrpc.New(ctx) // endpoint OTLP cấu hình qua env/secret
    if err != nil { return nil, err }
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(resource.NewSchemaless(
            semconv.ServiceName(serviceName))),
        sdktrace.WithSampler(errorBiasedSampler()), // §1 #10
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.TraceContext{}) // W3C traceparent (§1 #2,#8)
    return tp.Shutdown, nil
}
```

### Structured logging gắn trace_id (§1 #5, #6)

```go
// obs/logging.go
// Logger là slog JSON handler tự rút trace_id/request_id từ context và mask PII.
func FromContext(ctx context.Context) *slog.Logger {
    sc := trace.SpanContextFromContext(ctx)
    return base.With(
        slog.String("trace_id", sc.TraceID().String()),
        slog.String("request_id", requestIDFrom(ctx)),
    )
}

// Log với mask: gọi redact trên field nhạy cảm trước khi ghi.
func Info(ctx context.Context, msg string, fields ...slog.Attr) {
    FromContext(ctx).LogAttrs(ctx, slog.LevelInfo, msg, redactAttrs(fields)...) // §1 #6
}
```

### Middleware obs đầu chain (§1 #7, #8)

```go
// obs/middleware.go
func HTTP(serviceName string) func(http.Handler) http.Handler {
    tr := otel.Tracer(serviceName)
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Tiếp tục trace từ upstream qua traceparent; nếu chưa có, gốc từ X-Request-Id.
            ctx := otel.GetTextMapPropagator().Extract(r.Context(),
                propagation.HeaderCarrier(r.Header))
            ctx, span := tr.Start(ctx, r.Method+" "+routePattern(r))
            defer span.End()
            ctx = withRequestID(ctx, requestIDFromHeader(r))
            sw := &statusWriter{ResponseWriter: w}
            next.ServeHTTP(sw, r.WithContext(ctx))
            span.SetAttributes(semconv.HTTPStatusCode(sw.status))
            metrics.HTTPObserve(serviceName, routePattern(r), sw.status, sw.elapsed())
        })
    }
}
```

### Redact PII (§1 #6)

```go
// obs/redact.go
// redactAttrs mask field nhạy cảm theo tên; không bao giờ để email/phone/token thô.
func redactAttrs(in []slog.Attr) []slog.Attr {
    for i, a := range in {
        switch a.Key {
        case "email", "phone", "token", "cookie", "authorization":
            in[i] = slog.String(a.Key, mask(a.Value.String()))
        }
    }
    return in
}
```

---

## §4 - Acceptance criteria

1. Khởi tạo tracer cho một service -> `/metrics` phơi bày được; OTLP exporter cấu hình từ env/secret.
2. Request qua hai service giả (A gọi B) với traceparent -> cả hai span chung một `trace_id` (chuỗi nối liền).
3. Service B nhận request KHÔNG kèm traceparent từ A -> vẫn tạo span gốc từ `X-Request-Id` (không vỡ).
4. Mọi log line phát qua `obs.Info` -> là JSON hợp lệ VÀ chứa `trace_id` + `request_id`.
5. Log một sự kiện kèm field `email` -> giá trị email bị mask, KHÔNG xuất hiện thô trong output.
6. Log kèm `token`/`cookie` -> mask tương tự.
7. Metric `http_requests_total{service,route,status}` tăng đúng theo request; histogram `http_request_duration_ms` ghi nhận độ trễ.
8. Outgoing request tới downstream -> có header `traceparent` (inject thành công).
9. Endpoint `/metrics` không truy cập được từ public trần (chỉ nội bộ/auth).
10. Shutdown service -> span/metric còn trong buffer được flush (không mất tín hiệu cuối).
11. Grafana dashboard `overview.json` tồn tại trong repo và tham chiếu metric p95/error-rate/volume.
12. Dashboard hiển thị ngưỡng NFR-INFRA-001 (đường 300ms / 500ms) trên panel latency.

---

## §5 - Kiểm thử (verification)

```go
// obs/tracing_test.go
func TestTrace_ContinuesAcrossServices(t *testing.T) {
    // Service A khởi span, gọi B; B extract traceparent và tiếp tục.
    var aTrace, bTrace string
    srvB := httptest.NewServer(obs.HTTP("svc-b")(http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            bTrace = trace.SpanContextFromContext(r.Context()).TraceID().String()
        })))
    defer srvB.Close()

    srvA := httptest.NewServer(obs.HTTP("svc-a")(http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            aTrace = trace.SpanContextFromContext(r.Context()).TraceID().String()
            req, _ := http.NewRequestWithContext(r.Context(), "GET", srvB.URL, nil)
            otel.GetTextMapPropagator().Inject(r.Context(),
                propagation.HeaderCarrier(req.Header)) // §1 #8
            http.DefaultClient.Do(req)
        })))
    defer srvA.Close()

    http.Get(srvA.URL)
    require.NotEmpty(t, aTrace)
    require.Equal(t, aTrace, bTrace) // chung trace_id
}

// obs/logging_test.go
func TestLog_HasTraceAndRequestID(t *testing.T) {
    var buf bytes.Buffer
    ctx := ctxWithSpanAndRequestID(t, "req-123")
    obs.SetOutput(&buf)
    obs.Info(ctx, "tracked product")
    var m map[string]any
    require.NoError(t, json.Unmarshal(buf.Bytes(), &m)) // JSON hợp lệ
    require.NotEmpty(t, m["trace_id"])
    require.Equal(t, "req-123", m["request_id"])
}

// obs/redact_test.go
func TestLog_RedactsPII(t *testing.T) {
    var buf bytes.Buffer
    obs.SetOutput(&buf)
    obs.Info(ctx, "login", slog.String("email", "chi@example.com"),
        slog.String("token", "secret-jwt-xyz"))
    out := buf.String()
    require.NotContains(t, out, "chi@example.com")
    require.NotContains(t, out, "secret-jwt-xyz")
}

func TestMetrics_HTTPObserve(t *testing.T) {
    metrics.HTTPObserve("svc-a", "/v1/track", 200, 42*time.Millisecond)
    body := scrape(t, metricsHandler())
    require.Contains(t, body, `http_requests_total{route="/v1/track",service="svc-a",status="200"}`)
}
```

---

## §6 - Khung triển khai

Xem §3. Mỗi service: `InitTracer` lúc start, `HTTP(serviceName)` là middleware đầu chain (trước cả auth ở service nội bộ, hoặc ngay sau request-id ở gateway), `Shutdown` lúc dừng để flush. Log dùng slog JSON handler bọc redact. OTLP endpoint và Grafana datasource lấy cấu hình qua TASK-INFRA-003 (không nhúng). Dashboard JSON versioned trong repo để thay đổi qua review. Sampler thiên về lỗi: luôn giữ trace có span lỗi, sample tỉ lệ trace thành công.

---

## §7 - Phụ thuộc

- TASK-INFRA-001 - gateway cung cấp `X-Request-Id` làm gốc trace_id; obs middleware nối tiếp.
- TASK-INFRA-003 - OTLP endpoint, Grafana datasource creds lấy qua provider secret.
- TASK-COMPLY-004 (downstream) - quy trình breach 72h dựa vào trace/log/metric để phát hiện và điều tra.
- NFR-INFRA-001 - dashboard đối chiếu ngưỡng p95.
- Hạ tầng: OpenTelemetry Collector, Prometheus, Grafana.

---

## §8 - Payload ví dụ

### Một log line (JSON, đã mask)

```json
{"time":"2026-06-27T08:12:33Z","level":"INFO","msg":"tracked product","service":"track-svc","trace_id":"4bf92f3577b34da6a3ce929d0e0e4736","request_id":"6b1d-c004","user_id":90112,"email":"c***@e***.com"}
```

### Outgoing request có traceparent

```http
GET /v1/products/90112 HTTP/1.1
Host: price-svc.internal
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Log-based alerting (cảnh báo từ pattern log) ngoài metric-based - thêm khi có khối lượng log đủ.
- Exemplars nối metric <-> trace trực tiếp trên Grafana - tối ưu điều tra giai đoạn sau.
- Trace lấy mẫu thích ứng theo tải (head/tail sampling nâng cao) - khi chi phí trace thành vấn đề.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Trace đứt giữa chừng | test cross-service | mất chuỗi điều tra | Bắt buộc inject/extract traceparent (§1 #2,#8) |
| Log thiếu trace_id | logging_test | không join được log<->trace | Logger tự gắn từ context (§1 #5) |
| PII thô trong log | redact_test + review | vi phạm PDPL | Mask trước ghi (§1 #6) |
| /metrics public | scan/review | rò rỉ tín hiệu nội bộ | Bảo vệ nội bộ/auth (§1 #3) |
| Mất span/metric khi service dừng | test shutdown | thiếu tín hiệu cuối | Flush khi Shutdown (§1 #11) |
| OTLP collector down | metric exporter | mất trace tạm thời | Buffer + retry; metric vẫn pull được độc lập |
| Chi phí lưu trace bùng nổ | dashboard storage | tốn kém | Sampler thiên lỗi + tỉ lệ (§1 #10) |
| Nhãn metric high-cardinality | Prometheus load | OOM Prometheus | Dùng route pattern, không raw path/id |
| Đồng hồ service lệch | trace timeline lệch | span chồng sai | NTP đồng bộ; dùng thời gian monotonic cho duration |

---

## §11 - Ghi chú

- Ba trụ quan sát chỉ mạnh khi join được: metric ("cái gì sai") + trace ("ở chặng nào") + log ("chi tiết gì"), nối bằng một trace_id chung.
- Trace_id bắt nguồn từ `X-Request-Id` của gateway nối client tới đáy backend - một sợi chỉ xuyên suốt thao tác người dùng.
- W3C traceparent giữ chuỗi liền qua ranh giới HTTP; tạo trace_id mới giữa chừng là phá chuỗi.
- Mask PII trong log giữ log hữu ích để debug (vẫn có id ẩn danh) mà không biến log thành kho dữ liệu cá nhân vi phạm PDPL.
- Đưa ngưỡng NFR vào dashboard biến con số trừu tượng thành đường đỏ trực quan, phát hiện hồi quy p95 sớm.
- Đây là bề mặt điều tra cho breach 72h của TASK-COMPLY-004: phát hiện nhanh, lần vết nhanh.

---

*Hết TASK-INFRA-004. Status: ready_to_implement (mục tiêu audit 10/10).*
