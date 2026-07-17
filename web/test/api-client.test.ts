import { apiFetch } from "../lib/api";
import { setAccessToken, tryRefreshOnce, logout, getAccessToken } from "../lib/auth";

jest.mock("../lib/auth", () => ({
  setAccessToken: jest.fn(),
  getAccessToken: jest.fn(),
  tryRefreshOnce: jest.fn(),
  logout: jest.fn(),
}));

describe("API Client (lib/api.ts)", () => {
  let fetchMock: jest.Mock;

  beforeEach(() => {
    fetchMock = jest.fn();
    global.fetch = fetchMock;
    process.env.NEXT_PUBLIC_API_BASE_URL = "http://api.gateway";
    jest.clearAllMocks();
  });

  afterEach(() => {
    delete process.env.NEXT_PUBLIC_API_BASE_URL;
  });

  it("should attach Authorization header if access token exists", async () => {
    (getAccessToken as jest.Mock).mockReturnValue("fake-token");
    fetchMock.mockResolvedValue({ status: 200 });

    await apiFetch("/test");

    expect(fetchMock).toHaveBeenCalledWith(
      "http://api.gateway/test",
      expect.objectContaining({
        headers: expect.any(Headers),
      })
    );

    const callArgs = fetchMock.mock.calls[0];
    const headers = callArgs[1].headers as Headers;
    expect(headers.get("Authorization")).toBe("Bearer fake-token");
  });

  it("should refresh token once on 401 and retry", async () => {
    (getAccessToken as jest.Mock).mockReturnValueOnce("old-token").mockReturnValueOnce("new-token");

    // First call returns 401, second call returns 200
    fetchMock.mockResolvedValueOnce({ status: 401 }).mockResolvedValueOnce({ status: 200 });
    (tryRefreshOnce as jest.Mock).mockResolvedValue("refreshed");

    const res = await apiFetch("/protected");

    expect(tryRefreshOnce).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(res.status).toBe(200);

    const retryCallArgs = fetchMock.mock.calls[1];
    const retryHeaders = retryCallArgs[1].headers as Headers;
    expect(retryHeaders.get("Authorization")).toBe("Bearer new-token");
  });

  it("should logout and throw if refresh fails on 401", async () => {
    (getAccessToken as jest.Mock).mockReturnValue("old-token");
    fetchMock.mockResolvedValue({ status: 401 });
    (tryRefreshOnce as jest.Mock).mockResolvedValue("invalid");

    await expect(apiFetch("/protected")).rejects.toThrow("Unauthorized");

    expect(tryRefreshOnce).toHaveBeenCalledTimes(1);
    expect(logout).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledTimes(1); // Only the initial call
  });

  it("preserves the session on a transient refresh failure", async () => {
    (getAccessToken as jest.Mock).mockReturnValue("old-token");
    fetchMock.mockResolvedValue({ status: 401 });
    (tryRefreshOnce as jest.Mock).mockResolvedValue("transient");

    await expect(apiFetch("/protected")).rejects.toThrow("Authentication temporarily unavailable");

    expect(logout).not.toHaveBeenCalled();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
