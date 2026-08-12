/**
 * @jest-environment node
 */
import { GET as healthz } from "../app/api/healthz/route";
import { GET, POST } from "../app/v1/[...path]/route";
import { NextRequest } from "next/server";

describe("web healthz", () => {
  it("returns ok json", async () => {
    const res = await healthz();
    expect(res.status).toBe(200);
    await expect(res.json()).resolves.toEqual({ ok: true });
  });
});

describe("same-origin /v1 gateway proxy", () => {
  const originalFetch = global.fetch;
  const originalGateway = process.env.GATEWAY_INTERNAL_BASE_URL;

  afterEach(() => {
    global.fetch = originalFetch;
    if (originalGateway === undefined) delete process.env.GATEWAY_INTERNAL_BASE_URL;
    else process.env.GATEWAY_INTERNAL_BASE_URL = originalGateway;
  });

  it("404s /v1/auth so token-pair endpoints stay off the web origin", async () => {
    const req = new NextRequest("http://127.0.0.1:3000/v1/auth/login", { method: "POST" });
    const res = await POST(req, { params: Promise.resolve({ path: ["auth", "login"] }) });
    expect(res.status).toBe(404);
  });

  it("forwards allowlisted /v1 paths to the private gateway", async () => {
    process.env.GATEWAY_INTERNAL_BASE_URL = "http://gateway.internal:8080";
    const fetchMock = jest.fn().mockResolvedValue(
      new Response(JSON.stringify({ items: [] }), {
        status: 200,
        headers: { "Content-Type": "application/json", "X-Request-Id": "abc" },
      }),
    );
    global.fetch = fetchMock as unknown as typeof fetch;

    const req = new NextRequest("http://127.0.0.1:3000/v1/tracked-products", {
      headers: { Authorization: "Bearer access-token" },
    });
    const res = await GET(req, { params: Promise.resolve({ path: ["tracked-products"] }) });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("http://gateway.internal:8080/v1/tracked-products");
    expect((init.headers as Headers).get("authorization")).toBe("Bearer access-token");
    expect((init.headers as Headers).get("x-user-id")).toBeNull();
    expect(res.status).toBe(200);
    expect(res.headers.get("x-request-id")).toBe("abc");
  });

  it("does not forward operator or service tokens from the browser", async () => {
    process.env.GATEWAY_INTERNAL_BASE_URL = "http://gateway.internal:8080";
    const fetchMock = jest.fn().mockResolvedValue(new Response(null, { status: 204 }));
    global.fetch = fetchMock as unknown as typeof fetch;

    const req = new NextRequest("http://127.0.0.1:3000/v1/alerts", {
      method: "GET",
      headers: {
        Authorization: "Bearer access-token",
        "X-Service-Token": "stolen-service",
        "X-Operator-Token": "stolen-operator",
        "X-User-Id": "999",
      },
    });
    await GET(req, { params: Promise.resolve({ path: ["alerts"] }) });
    const headers = fetchMock.mock.calls[0][1].headers as Headers;
    expect(headers.get("x-service-token")).toBeNull();
    expect(headers.get("x-operator-token")).toBeNull();
    expect(headers.get("x-user-id")).toBeNull();
    expect(headers.get("authorization")).toBe("Bearer access-token");
  });

  it("returns 502 when the gateway is unreachable", async () => {
    process.env.GATEWAY_INTERNAL_BASE_URL = "http://gateway.internal:8080";
    global.fetch = jest.fn().mockRejectedValue(new Error("connect")) as unknown as typeof fetch;
    const req = new NextRequest("http://127.0.0.1:3000/v1/alerts");
    const res = await GET(req, { params: Promise.resolve({ path: ["alerts"] }) });
    expect(res.status).toBe(502);
  });
});
