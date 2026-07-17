/**
 * @jest-environment node
 */
import { NextRequest } from "next/server";

jest.mock("next/headers", () => ({ cookies: jest.fn() }));
jest.mock("../lib/server-auth", () => ({
  gatewayFetch: jest.fn(),
  readRefreshCookie: jest.fn(),
  refreshCookieName: "__Host-sandeal_refresh",
  refreshCookieOptions: { httpOnly: true, secure: true, sameSite: "strict", path: "/", maxAge: 1 },
  requestHasAllowedOrigin: jest.fn(),
}));

import { POST } from "../app/api/auth/refresh/route";

const { cookies: mockCookies } = jest.requireMock("next/headers") as { cookies: jest.Mock };
const {
  gatewayFetch: mockGatewayFetch,
  readRefreshCookie: mockReadRefreshCookie,
  requestHasAllowedOrigin: mockRequestHasAllowedOrigin,
} = jest.requireMock("../lib/server-auth") as Record<string, jest.Mock>;

function refreshRequest() {
  return new NextRequest("https://sandeal.example/api/auth/refresh", {
    method: "POST",
    headers: { origin: "https://sandeal.example" },
  });
}

describe("POST /api/auth/refresh", () => {
  const cookieStore = { delete: jest.fn(), set: jest.fn() };

  beforeEach(() => {
    jest.clearAllMocks();
    mockCookies.mockReturnValue(cookieStore);
    mockRequestHasAllowedOrigin.mockReturnValue(true);
    mockReadRefreshCookie.mockReturnValue("refresh-token");
  });

  it.each([429, 502, 503])("keeps the cookie on transient upstream status %i", async (status) => {
    mockGatewayFetch.mockResolvedValue({ ok: false, status } as Response);
    const request = refreshRequest();

    const response = await POST(request);

    expect(response.status).toBe(status);
    expect(cookieStore.delete).not.toHaveBeenCalled();
    expect(mockGatewayFetch.mock.calls[0][2]).toBe(request);
  });

  it("clears all refresh cookie names only for an invalid refresh token", async () => {
    mockGatewayFetch.mockResolvedValue({ ok: false, status: 401 } as Response);

    const response = await POST(refreshRequest());

    expect(response.status).toBe(401);
    expect(cookieStore.delete).toHaveBeenCalledWith("__Host-sandeal_refresh");
    expect(cookieStore.delete).toHaveBeenCalledWith("sandeal_refresh");
    expect(cookieStore.delete).toHaveBeenCalledWith("refresh_token");
  });

  it("keeps the cookie when the gateway request itself fails", async () => {
    mockGatewayFetch.mockRejectedValue(new Error("gateway unavailable"));

    const response = await POST(refreshRequest());

    expect(response.status).toBe(503);
    expect(cookieStore.delete).not.toHaveBeenCalled();
  });
});
