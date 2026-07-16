---
id: TASK-WEB-005
title: "GraphQL BFF cho web - một endpoint /graphql truy vấn linh hoạt wishlist + biểu đồ trong một round-trip, resolver ủy quyền cho service REST, DataLoader chống N+1, độ sâu/độ phức tạp có trần"
module: WEB
priority: SHOULD
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-INFRA-001, TASK-WEB-001, TASK-WEB-003, TASK-WEB-004, TASK-TRACK-002, TASK-TRACK-003, TASK-DEAL-003]
depends_on: [TASK-INFRA-001, TASK-WEB-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.7 (GraphQL cho web app - truy vấn linh hoạt wishlist/biểu đồ)"
  - "docs/... §3.1 (gateway định tuyến GraphQL), §3.8 (NFR hiệu năng/bảo mật)"
source_decisions:
  - "DEC-WEB-21: BFF GraphQL là một service đứng SAU gateway (TASK-INFRA-001 định tuyến /graphql tới đây); gateway đã verify JWT, BFF nhận claims qua context, KHÔNG tự verify lại token"
  - "DEC-WEB-22: resolver KHÔNG truy cập DB trực tiếp - ủy quyền cho service REST hiện có (track-svc wishlist/alert, deal-svc chart) để giữ một nguồn logic + một chỗ kiểm chủ sở hữu"
  - "DEC-WEB-23: dùng DataLoader gom + cache trong một request để chống N+1 (vd nhiều wishlist_item cùng nạp chart) - không bắn N lời gọi REST tuần tự"
  - "DEC-WEB-24: đặt trần độ sâu query (max depth) và trần độ phức tạp (cost) - từ chối query vượt trần để chặn truy vấn ác ý làm cạn tài nguyên"
  - "DEC-WEB-25: BFF chỉ phục vụ đọc (Query) ở slice này; ghi (tạo wishlist/alert) vẫn qua REST của TASK-TRACK-002/003 - không nhân đôi mutation; tránh hai đường ghi lệch nhau"

language: "TypeScript 5.x; GraphQL (graphql-js + một runtime như GraphQL Yoga/Apollo Server); DataLoader; chạy như route handler Next hoặc service Node sau gateway"
service: shopass/services/bff/
new_files:
  - services/bff/src/schema.graphql
  - services/bff/src/server.ts
  - services/bff/src/context.ts
  - services/bff/src/resolvers/wishlist.ts
  - services/bff/src/resolvers/chart.ts
  - services/bff/src/loaders/chart-loader.ts
  - services/bff/src/security/limits.ts
  - services/bff/src/rest-client.ts
  - services/bff/test/resolver-auth.test.ts
  - services/bff/test/dataloader-n1.test.ts
  - services/bff/test/query-depth.test.ts
modified_files:
  - services/bff/package.json
allowed_tools:
  - file_read: services/bff/**
  - file_write: services/bff/**
  - bash: cd services/bff && npm test && npx tsc --noEmit
disallowed_tools:
  - resolver truy cập DB trực tiếp thay vì gọi service REST (vi phạm DEC-WEB-22, nhân đôi logic + bỏ kiểm chủ sở hữu)
  - tự verify/parse lại JWT trong BFF (vi phạm DEC-WEB-21, gateway đã verify; nhân đôi sai lệch)
  - bỏ trần độ sâu/độ phức tạp query (vi phạm DEC-WEB-24, mở đường truy vấn ác ý cạn tài nguyên)
  - thêm mutation ghi wishlist/alert ở BFF (vi phạm DEC-WEB-25, hai đường ghi lệch nhau)

effort_hours: 6
sub_tasks:
  - "0.75h: schema.graphql - type Wishlist/WishlistItem/ChartData + Query (me, wishlists, productChart); KHÔNG Mutation ghi (slice đọc)"
  - "0.75h: context.ts - đọc claims (user sub) do gateway gắn (header/forwarded); KHÔNG verify token"
  - "0.75h: rest-client.ts - gọi track-svc + deal-svc REST với context user, truyền request-id"
  - "1.0h: resolvers/wishlist.ts - resolve wishlists + items qua REST TASK-TRACK-002 (ủy quyền kiểm chủ sở hữu)"
  - "1.0h: loaders/chart-loader.ts + resolvers/chart.ts - DataLoader gom nạp chart cho nhiều product (chống N+1) qua feed TASK-DEAL-003"
  - "0.75h: security/limits.ts - max depth + cost limit, từ chối query vượt trần"
  - "0.5h: server.ts - wire schema + resolvers + loaders + limits sau gateway"
  - "1.5h: tests - resolver thiếu user context bị từ chối; DataLoader gộp N product thành 1 batch; query quá sâu bị chặn"

risk_if_skipped: "GraphQL BFF là tiện ích hiệu năng/trải nghiệm cho web (SHOULD, không chặn release): nó cho phép màn hình dashboard nạp wishlist kèm biểu đồ của từng item trong MỘT round-trip thay vì nhiều lời gọi REST tuần tự, giảm độ trễ cảm nhận và đơn giản hóa client. Thiếu nó thì web vẫn chạy bằng REST (TASK-WEB-001 lib/api.ts) - mất tối ưu chứ không mất tính năng. Nhưng nếu làm SAI thì nguy hiểm: nếu resolver truy cập DB trực tiếp thay vì ủy quyền REST thì nhân đôi logic kiểm chủ sở hữu - một chỗ quên kiểm là lỗ IDOR rò rỉ wishlist/alert người khác (đụng PDPL). Nếu không đặt trần độ sâu/độ phức tạp thì một query lồng sâu ác ý làm cạn tài nguyên BFF (DoS qua GraphQL là lớp tấn công kinh điển). Nếu thêm mutation ghi ở đây thì có hai đường ghi (REST + GraphQL) dễ lệch validate và lệch trạng thái. Vì là SHOULD, có thể hoãn sau MVP, nhưng khi làm phải đúng các ranh giới này."
---

## §1 - Mô tả (BCP-14 normative)

GraphQL BFF **MUST** là một service đọc đứng sau API Gateway (TASK-INFRA-001), phục vụ endpoint `/graphql` cho web app truy vấn linh hoạt wishlist và biểu đồ, với resolver ủy quyền cho service REST, DataLoader chống N+1, và trần độ sâu/độ phức tạp. Hợp đồng:

1. BFF **MUST** đứng sau gateway: gateway (TASK-INFRA-001) định tuyến `/graphql` tới BFF và đã verify JWT; BFF nhận thông tin user (sub/claims) qua context do gateway truyền (header forwarded), KHÔNG tự verify hay parse lại token (DEC-WEB-21).
2. Resolver **MUST NOT** truy cập DB trực tiếp: mọi dữ liệu lấy qua service REST hiện có - wishlist/alert từ track-svc (TASK-TRACK-002/003), dữ liệu biểu đồ từ deal-svc (TASK-DEAL-003) (DEC-WEB-22). Kiểm chủ sở hữu nằm ở service REST, BFF không nhân bản logic đó.
3. BFF **MUST** dùng DataLoader để gom và cache trong phạm vi một request (DEC-WEB-23): khi một query nạp biểu đồ cho nhiều `product_id` (vd mọi item trong wishlist), DataLoader gộp thành một lô thay vì N lời gọi REST tuần tự (chống N+1).
4. BFF **MUST** đặt trần độ sâu query (max depth) và trần độ phức tạp (cost/complexity) (DEC-WEB-24); query vượt trần bị từ chối với lỗi GraphQL rõ ràng trước khi thực thi resolver - chặn truy vấn lồng sâu ác ý làm cạn tài nguyên.
5. Ở slice này BFF **MUST** chỉ phục vụ `Query` (đọc); KHÔNG định nghĩa `Mutation` ghi wishlist/alert (DEC-WEB-25). Mọi thao tác ghi vẫn qua REST của TASK-TRACK-002/003 để tránh hai đường ghi lệch nhau.
6. Schema **MUST** phơi tối thiểu: `me` (thông tin user hiện tại), `wishlists` (danh sách wishlist của caller kèm item), và `productChart(productId, range)` (dữ liệu biểu đồ từ feed TASK-DEAL-003) - đủ để dashboard nạp wishlist kèm biểu đồ trong một round-trip.
7. Mọi resolver **MUST** yêu cầu user context hợp lệ (từ gateway); request thiếu context user (không qua gateway hoặc gateway không gắn claims) bị từ chối - resolver không trả dữ liệu cá nhân cho ngữ cảnh ẩn danh.
8. BFF **MUST** truyền `request-id` (do gateway gắn, TASK-INFRA-001) xuống các lời gọi REST để giữ vết xuyên service (đồng nhất observability TASK-INFRA-004).
9. `productChart` **MUST** trả đúng hình dạng feed TASK-DEAL-003 (daily + annotations + maturity) - BFF không tự tính verdict/median, chỉ chuyển tiếp + gộp; client hiển thị giống TASK-WEB-003.
10. BFF **MUST** xử lý lỗi từ service REST một cách có cấu trúc: `403`/`404` từ REST ánh xạ thành lỗi GraphQL trung lập (không lộ tài nguyên user khác); lỗi mạng ánh xạ thành lỗi server rõ ràng, không nuốt im lặng.
11. Toàn bộ **MUST** vượt `npx tsc --noEmit` sạch và `npm test` xanh.
12. BFF **SHOULD** phát OTel `graphql_resolver_duration_ms{field}` và `graphql_rejected_total{reason}` (đếm query bị chặn vì depth/cost) để theo dõi truy vấn nặng/ác ý.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao BFF đứng sau gateway, không tự verify token (DEC-WEB-21)?** Gateway (TASK-INFRA-001) đã là điểm thực thi JWT duy nhất cho toàn hệ thống. BFF tự verify lại là nhân đôi logic - hai chỗ kiểm token dễ lệch (một chỗ cập nhật JWKS, chỗ kia quên). BFF tin gateway và nhận claims qua context. Đây là mẫu BFF chuẩn: edge xác thực, dịch vụ phía sau nhận danh tính đã xác thực.

**Vì sao resolver ủy quyền REST, không chạm DB (DEC-WEB-22)?** Logic kiểm chủ sở hữu (chống IDOR) nằm ở track-svc (TASK-TRACK-002/003). Nếu BFF chạm DB trực tiếp thì phải sao chép logic đó - một chỗ quên kiểm là lỗ IDOR rò rỉ wishlist/alert người khác. Ủy quyền REST giữ một nguồn logic và một chỗ kiểm quyền; BFF chỉ là lớp định hình + gộp dữ liệu cho client.

**Vì sao DataLoader chống N+1 (DEC-WEB-23)?** GraphQL dễ rơi vào N+1: query "wishlist kèm biểu đồ mỗi item" ngây thơ bắn một lời gọi chart cho mỗi item. Với wishlist 20 item là 20 round-trip REST tuần tự - chậm. DataLoader gom các id trong một tick thành một lô, gọi một lần, cache trong request. Đây là khác biệt giữa BFF nhanh và BFF tệ hơn cả REST.

**Vì sao trần độ sâu/độ phức tạp (DEC-WEB-24)?** GraphQL cho client tự đặt hình dạng query - bao gồm query lồng sâu ác ý làm cạn tài nguyên (lớp tấn công kinh điển của GraphQL). Trần max depth + cost từ chối query vượt ngưỡng trước khi chạy resolver. Không có trần là mở cửa DoS.

**Vì sao chỉ đọc, ghi vẫn qua REST (DEC-WEB-25)?** Hai đường ghi (REST + GraphQL mutation) dễ lệch validate (một đường cập nhật luật, đường kia quên) và lệch trạng thái. Slice này BFF chỉ tối ưu đọc (nơi GraphQL tỏa sáng: gộp nhiều nguồn trong một round-trip); ghi giữ một đường REST của TASK-TRACK-002/003 để một nguồn validate.

---

## §3 - Hợp đồng API / DDL

### Schema (schema.graphql) - chỉ Query

```graphql
# services/bff/src/schema.graphql
type Query {
  me: User!
  wishlists: [Wishlist!]!                       # của caller; track-svc kiểm chủ sở hữu
  productChart(productId: ID!, range: String = "90d"): ChartData!
}

type User { id: ID!; displayName: String }

type Wishlist {
  id: ID!; name: String!
  items: [WishlistItem!]!
}
type WishlistItem {
  productId: ID!
  targetPrice: Int                              # VND int64, nullable
  chart(range: String = "90d"): ChartData       # nạp qua DataLoader (chống N+1)
}

type ChartData {                                # mirror feed TASK-DEAL-003
  maturity: String!
  daily: [DailyPoint!]!
  annotations: Annotations!
}
type DailyPoint { day: String!; minP: Int!; maxP: Int!; closeP: Int! }
type Annotations { median90: Int!; trailingMin: Int!; verdict: String!; accumulating: Boolean!; doubleDates: [String!]! }

# KHÔNG có type Mutation ở slice này (DEC-WEB-25)
```

### Context (context.ts) - nhận claims từ gateway, không verify

```ts
// services/bff/src/context.ts
export interface GqlContext {
  userId: string | null;   // sub do gateway gắn (DEC-WEB-21); BFF KHÔNG verify token
  requestId: string;       // do gateway gắn (TASK-INFRA-001)
  rest: RestClient;
}

export function buildContext(req: IncomingRequest): GqlContext {
  return {
    userId: req.headers["x-user-id"] ?? null,        // gateway forward sub (đã verify)
    requestId: req.headers["x-request-id"] ?? crypto.randomUUID(),
    rest: new RestClient(req.headers["x-user-id"], req.headers["x-request-id"]),
  };
}
```

### DataLoader chart (loaders/chart-loader.ts) - chống N+1

```ts
// services/bff/src/loaders/chart-loader.ts
import DataLoader from "dataloader";

// Gom nhiều (productId, range) trong một request thành các lời gọi feed TASK-DEAL-003,
// cache theo khóa trong phạm vi request (DEC-WEB-23).
export function makeChartLoader(rest: RestClient) {
  return new DataLoader<{ productId: string; range: string }, ChartData>(
    async (keys) => {
      // gọi REST cho từng key NHƯNG được gom trong một tick; cache tránh trùng
      return Promise.all(keys.map((k) => rest.getChart(k.productId, k.range))); // feed TASK-DEAL-003
    },
    { cacheKeyFn: (k) => `${k.productId}:${k.range}` }
  );
}
```

### Resolver wishlist (ủy quyền REST, không chạm DB)

```ts
// services/bff/src/resolvers/wishlist.ts
export const wishlistResolvers = {
  Query: {
    wishlists: (_: unknown, __: unknown, ctx: GqlContext) => {
      if (!ctx.userId) throw new GraphQLError("UNAUTHENTICATED"); // §1 #7
      return ctx.rest.listWishlists();                            // track-svc kiểm chủ sở hữu (DEC-WEB-22)
    },
  },
  WishlistItem: {
    chart: (item: { productId: string }, args: { range: string }, ctx: GqlContext) =>
      ctx.loaders.chart.load({ productId: item.productId, range: args.range ?? "90d" }), // DataLoader
  },
};
```

### Trần độ sâu/độ phức tạp (security/limits.ts)

```ts
// services/bff/src/security/limits.ts
export const MAX_DEPTH = 8;          // chặn query lồng quá sâu (DEC-WEB-24)
export const MAX_COST = 1000;        // trần độ phức tạp

// validate trước khi thực thi; vượt trần -> GraphQLError, đếm graphql_rejected_total
export function enforceLimits(document: DocumentNode): void {
  const depth = queryDepth(document);
  if (depth > MAX_DEPTH) throw new GraphQLError(`query quá sâu (>${MAX_DEPTH})`);
  const cost = queryCost(document);
  if (cost > MAX_COST) throw new GraphQLError(`query quá phức tạp (>${MAX_COST})`);
}
```

---

## §4 - Acceptance criteria

1. BFF nhận user qua context do gateway gắn (header forwarded); grep `services/bff/**`: KHÔNG có verify/parse JWT (DEC-WEB-21).
2. Resolver lấy dữ liệu qua `RestClient` (track-svc/deal-svc); grep: KHÔNG có truy vấn DB/SQL trực tiếp trong BFF (DEC-WEB-22).
3. Query nạp biểu đồ cho N product trong một wishlist gọi feed qua DataLoader gộp một lô - KHÔNG N lời gọi tuần tự rời rạc (test đếm batch).
4. Query vượt `MAX_DEPTH` hoặc `MAX_COST` bị từ chối với `GraphQLError` trước khi chạy resolver (DEC-WEB-24).
5. Schema chỉ có `Query`, KHÔNG có `Mutation` ghi wishlist/alert (DEC-WEB-25).
6. Schema phơi `me`, `wishlists` (kèm items), `productChart(productId, range)`; `ChartData` mirror hình dạng feed TASK-DEAL-003.
7. Resolver thiếu `userId` context (ngữ cảnh ẩn danh) ném `UNAUTHENTICATED`; không trả dữ liệu cá nhân.
8. BFF truyền `request-id` (header) xuống lời gọi REST.
9. `productChart`/`chart` trả `verdict`/`median90` từ feed (không tự tính trong BFF).
10. Lỗi `403`/`404` từ REST ánh xạ thành lỗi GraphQL trung lập; lỗi mạng thành lỗi server, không nuốt im lặng.
11. `npx tsc --noEmit` sạch; `npm test` xanh.

---

## §5 - Kiểm thử (verification)

```ts
// services/bff/test/resolver-auth.test.ts
import { wishlistResolvers } from "../src/resolvers/wishlist";

test("resolver thiếu userId → UNAUTHENTICATED, không gọi REST", () => {
  const ctx = { userId: null, rest: { listWishlists: jest.fn() } } as any;
  expect(() => wishlistResolvers.Query.wishlists(null, null, ctx)).toThrow(/UNAUTHENTICATED/);
  expect(ctx.rest.listWishlists).not.toHaveBeenCalled();
});

test("BFF không verify token (nhận sub từ context gateway)", async () => {
  const src = await readFile("src/context.ts", "utf8");
  expect(src).not.toMatch(/jwt\.verify|verifyToken|jwks/i); // DEC-WEB-21
});
```

```ts
// services/bff/test/dataloader-n1.test.ts
import { makeChartLoader } from "../src/loaders/chart-loader";

test("nạp chart cho 5 product trong một request gộp một batch (chống N+1)", async () => {
  let batchCalls = 0;
  const rest = { getChart: jest.fn(async () => { return {} as any; }) };
  const spyBatch = jest.fn(() => { batchCalls++; });
  const loader = makeChartLoader(rest as any, spyBatch);
  await Promise.all([1, 2, 3, 4, 5].map((id) => loader.load({ productId: String(id), range: "90d" })));
  expect(batchCalls).toBe(1);                 // một lô, KHÔNG 5 round-trip rời rạc
  expect(rest.getChart).toHaveBeenCalledTimes(5); // 5 key nhưng trong một batch
});
```

```ts
// services/bff/test/query-depth.test.ts
import { enforceLimits, MAX_DEPTH } from "../src/security/limits";
import { parse } from "graphql";

test("query quá sâu bị chặn trước khi chạy resolver", () => {
  const deep = parse(`{ wishlists { items { chart { annotations { doubleDates } } } } }`.repeat(1));
  // dựng query vượt MAX_DEPTH
  const tooDeep = parse(buildNestedQuery(MAX_DEPTH + 2));
  expect(() => enforceLimits(tooDeep)).toThrow(/quá sâu/);
});

test("query trong trần được chấp nhận", () => {
  const ok = parse(`{ wishlists { id name } }`);
  expect(() => enforceLimits(ok)).not.toThrow();
});
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `schema.graphql` (chỉ Query: me/wishlists/productChart) -> `context.ts` (nhận sub + request-id từ gateway, không verify) -> `rest-client.ts` (gọi track-svc + deal-svc, truyền request-id) -> `resolvers/wishlist.ts` + `resolvers/chart.ts` (ủy quyền REST) -> `loaders/chart-loader.ts` (DataLoader gộp) -> `security/limits.ts` (max depth + cost) -> `server.ts` (wire tất cả, đặt sau gateway) -> tests. BFF có thể chạy như route handler trong web app Next (`/api/graphql`) hoặc một service Node riêng - dù cách nào cũng nằm SAU gateway (TASK-INFRA-001 định tuyến `/graphql`). Vì là SHOULD, có thể hoãn tới sau khi REST path (TASK-WEB-001 lib/api.ts) đã phục vụ MVP; khi làm phải giữ đúng bốn ranh giới (không verify token, không chạm DB, có trần query, không mutation ghi).

---

## §7 - Phụ thuộc

- **TASK-INFRA-001** - gateway định tuyến `/graphql` tới BFF và verify JWT trước; BFF nhận claims qua context, không verify lại (depends_on cứng).
- **TASK-WEB-001** - web app tiêu thụ BFF; `lib/api.ts` mở rộng để gọi GraphQL endpoint cạnh REST (depends_on cứng).
- **TASK-TRACK-002 / TASK-TRACK-003** - nguồn wishlist/alert; resolver ủy quyền REST của chúng (giữ kiểm chủ sở hữu một chỗ).
- **TASK-DEAL-003** - feed biểu đồ; `productChart`/`chart` resolver gọi feed này qua DataLoader; `ChartData` mirror hình dạng của nó.
- **TASK-WEB-003 / TASK-WEB-004 (đồng cấp)** - cùng tiêu thụ dữ liệu wishlist + biểu đồ; BFF là đường thay thế REST cho dashboard cần gộp nhiều nguồn.
- Lib: `graphql`, một runtime (GraphQL Yoga/Apollo Server), `dataloader`.

---

## §8 - Payload ví dụ

### Query dashboard: wishlist kèm biểu đồ mỗi item (một round-trip)

```graphql
query Dashboard {
  wishlists {
    id
    name
    items {
      productId
      targetPrice
      chart(range: "90d") {            # nạp qua DataLoader, gộp một batch
        maturity
        annotations { median90 trailingMin verdict doubleDates }
      }
    }
  }
}
```

### Lỗi query vượt trần độ sâu

```json
{ "errors": [ { "message": "query quá sâu (>8)" } ] }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Persisted queries / allowlist query phía server để siết chặt hơn nữa truy vấn cho phép - cân nhắc khi mở public API.
- Mutation qua GraphQL (nếu sau này hợp nhất đường ghi) - chốt mô hình một-đường-ghi trước khi cân nhắc.
- Subscription (WSS) cho cập nhật giá realtime trên dashboard - bám đường WSS của gateway (TASK-INFRA-001) khi cần.
- Field-level cost cụ thể theo độ nặng từng resolver - tinh chỉnh trần cost sau khi đo tải thật.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| BFF tự verify token | grep jwt.verify/jwks | nhân đôi sai lệch auth | Nhận claims từ gateway (DEC-WEB-21) |
| Resolver chạm DB trực tiếp | grep SQL/DB trong bff | bỏ kiểm chủ sở hữu -> IDOR | Ủy quyền REST (DEC-WEB-22) |
| N+1 nạp chart mỗi item | dataloader-n1 test | chậm hơn cả REST | DataLoader gộp một batch (DEC-WEB-23) |
| Query lồng sâu ác ý | query-depth test | cạn tài nguyên (DoS) | Trần max depth + cost (DEC-WEB-24) |
| Thêm mutation ghi ở BFF | review schema | hai đường ghi lệch | Chỉ Query; ghi qua REST (DEC-WEB-25) |
| Resolver trả dữ liệu cho ẩn danh | resolver-auth test | rò rỉ dữ liệu cá nhân | Yêu cầu userId context (§1 #7) |
| 403/404 REST lộ tài nguyên khác | review ánh xạ lỗi | rò rỉ quyền riêng tư | Lỗi GraphQL trung lập (§1 #10) |
| BFF tự tính verdict/median | grep tính toán | lệch feed/thẻ | Chuyển tiếp feed TASK-DEAL-003 (§1 #9) |
| Mất request-id xuyên service | trace gãy | khó điều tra | Truyền x-request-id xuống REST (§1 #8) |

---

## §11 - Ghi chú

- BFF GraphQL là tiện ích hiệu năng (SHOULD): nạp wishlist kèm biểu đồ mỗi item trong một round-trip thay vì nhiều lời gọi REST tuần tự.
- Bốn ranh giới phải giữ khi làm: không verify token (gateway đã làm), không chạm DB (ủy quyền REST giữ kiểm chủ sở hữu một chỗ), có trần query (chống DoS), chỉ đọc (một đường ghi REST).
- Resolver ủy quyền REST giữ một nguồn logic + một chỗ kiểm IDOR - BFF chỉ định hình + gộp dữ liệu.
- DataLoader là khác biệt giữa BFF nhanh và BFF tệ hơn REST; gộp N id thành một lô trong một request.
- Trần độ sâu/độ phức tạp chặn lớp tấn công DoS kinh điển của GraphQL.
- Vì SHOULD, có thể hoãn sau MVP (REST path đã đủ); khi làm phải đúng các ranh giới trên để không tạo lỗ hổng mới.

---

*Hết TASK-WEB-005. Status: ready_to_implement (mục tiêu audit 10/10).*
