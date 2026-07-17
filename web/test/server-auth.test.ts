/**
 * @jest-environment node
 */
import { NextRequest } from "next/server";
import {
  gatewayFetch,
  readRefreshCookie,
  refreshCookieCandidates,
} from "../lib/server-auth";

describe("server auth helpers", () => {
  const originalFetch = global.fetch;

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it("never accepts legacy refresh cookies in production", () => {
    expect(refreshCookieCandidates("production")).toEqual(["__Host-sandeal_refresh"]);

    const legacyOnly = new NextRequest("https://sandeal.example/api/auth/refresh");
    legacyOnly.cookies.set("refresh_token", "legacy-token");
    expect(readRefreshCookie(legacyOnly, "production")).toBeUndefined();

    const current = new NextRequest("https://sandeal.example/api/auth/refresh");
    current.cookies.set("refresh_token", "legacy-token");
    current.cookies.set("__Host-sandeal_refresh", "host-token");
    expect(readRefreshCookie(current, "production")).toBe("host-token");
  });

  it("retains legacy compatibility only outside production", () => {
    expect(refreshCookieCandidates("development")).toEqual(["sandeal_refresh", "refresh_token"]);
    const request = new NextRequest("http://localhost/api/auth/refresh");
    request.cookies.set("refresh_token", "legacy-token");
    expect(readRefreshCookie(request, "development")).toBe("legacy-token");
  });

  it("forwards only Caddy's inbound X-Real-IP to the gateway", async () => {
    const fetchMock = jest.fn().mockResolvedValue({ ok: true, status: 204 } as Response);
    global.fetch = fetchMock;
    const request = new NextRequest("https://sandeal.example/api/auth/login", {
      headers: { "X-Real-IP": "203.0.113.10" },
    });

    await gatewayFetch(
      "/v1/auth/login",
      { headers: { "Content-Type": "application/json", "X-Real-IP": "forged" } },
      request,
    );

    const headers = fetchMock.mock.calls[0][1].headers as Headers;
    expect(headers.get("X-Real-IP")).toBe("203.0.113.10");
    expect(headers.get("Content-Type")).toBe("application/json");
  });

  it("does not relay a RequestInit-supplied X-Real-IP without an inbound request", async () => {
    const fetchMock = jest.fn().mockResolvedValue({ ok: true, status: 204 } as Response);
    global.fetch = fetchMock;

    await gatewayFetch("/v1/auth/login", { headers: { "X-Real-IP": "forged" } });

    const headers = fetchMock.mock.calls[0][1].headers as Headers;
    expect(headers.get("X-Real-IP")).toBeNull();
  });
});
