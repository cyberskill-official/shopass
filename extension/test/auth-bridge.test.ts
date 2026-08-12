import { authedFetch, setJwt, refreshJwt, NoAuthError } from "../src/sync/auth-bridge";
import { SyncEnvelope } from "../src/shared/types";

describe("auth-bridge", () => {
  beforeEach(() => {
    // mock fetch
    global.fetch = jest.fn(() =>
      Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve({}),
      })
    ) as jest.Mock;
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it("throws NoAuthError if no JWT", async () => {
    await setJwt(undefined);
    await expect(authedFetch("http://localhost", { payload: { platform: "shopee", items: [], vouchers: [] }, clientTs: 1 })).rejects.toThrow(NoAuthError);
  });

  it("adds Authorization Bearer header if JWT present", async () => {
    await setJwt("fake-jwt-token");
    await authedFetch("http://localhost", { payload: { platform: "shopee", items: [], vouchers: [] }, clientTs: 1 });
    
    expect(global.fetch).toHaveBeenCalledWith("http://localhost", expect.objectContaining({
      method: "POST",
      headers: {
        "Authorization": "Bearer fake-jwt-token",
        "Content-Type": "application/json"
      }
    }));
  });

  it("refreshJwt does not mint a token", async () => {
    await setJwt(undefined);
    await refreshJwt();
    await expect(
      authedFetch("http://localhost", {
        payload: { platform: "shopee", items: [], vouchers: [] },
        clientTs: 1,
      }),
    ).rejects.toThrow(NoAuthError);
    expect(global.fetch).not.toHaveBeenCalled();
  });
});
