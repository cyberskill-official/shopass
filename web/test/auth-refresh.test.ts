import { getAccessToken, setAccessToken, tryRefreshOnce } from "../lib/auth";

describe("client refresh", () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    setAccessToken(null);
  });

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it("coalesces concurrent refresh attempts into one request", async () => {
    let resolveResponse: (response: Response) => void = () => undefined;
    const pendingResponse = new Promise<Response>((resolve) => {
      resolveResponse = resolve;
    });
    const fetchMock = jest.fn().mockReturnValue(pendingResponse);
    global.fetch = fetchMock;

    const first = tryRefreshOnce();
    const second = tryRefreshOnce();

    expect(fetchMock).toHaveBeenCalledTimes(1);
    resolveResponse({
      ok: true,
      status: 200,
      json: async () => ({ accessToken: "fresh-token" }),
    } as Response);

    await expect(Promise.all([first, second])).resolves.toEqual(["refreshed", "refreshed"]);
    expect(getAccessToken()).toBe("fresh-token");
  });

  it("classifies a failed refresh as transient instead of leaking an exception", async () => {
    global.fetch = jest.fn().mockRejectedValue(new Error("network unavailable"));

    await expect(tryRefreshOnce()).resolves.toBe("transient");
  });
});
