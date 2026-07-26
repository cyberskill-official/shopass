import { AuthClient } from "./authClient";
import { KeychainSecureStore, MemoryAccessToken, MemorySecureStore } from "./tokenStore";
import { __resetMockKeychain } from "./__mocks__/react-native-keychain";

describe("authClient + tokenStore", () => {
  beforeEach(() => {
    __resetMockKeychain();
  });

  it("stores refresh in keychain-backed secure store, access in memory", async () => {
    const secure = new KeychainSecureStore();
    const access = new MemoryAccessToken();
    const fetchImpl = (async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url.endsWith("/v1/auth/login")) {
        return new Response(
          JSON.stringify({ access_token: "a1", refresh_token: "r1" }),
          { status: 200 },
        );
      }
      throw new Error(`unexpected ${url} ${init?.method}`);
    }) as typeof fetch;

    const auth = new AuthClient({
      baseURL: "https://api.test",
      secure,
      access,
      fetchImpl,
    });
    await auth.login("a@b.c", "pw");
    expect(access.get()).toBe("a1");
    expect(await secure.getRefresh()).toBe("r1");
  });

  it("clears tokens on logout", async () => {
    const secure = new MemorySecureStore();
    const access = new MemoryAccessToken();
    access.set("a");
    await secure.setRefresh("r");
    let unregistered = false;
    const auth = new AuthClient({
      baseURL: "https://api.test",
      secure,
      access,
      fetchImpl: fetch,
      onLogoutDevice: async () => {
        unregistered = true;
      },
    });
    await auth.logout();
    expect(access.get()).toBeNull();
    expect(await secure.getRefresh()).toBeNull();
    expect(unregistered).toBe(true);
  });

  it("refreshOnce updates access token", async () => {
    const secure = new MemorySecureStore();
    await secure.setRefresh("r0");
    const access = new MemoryAccessToken();
    const fetchImpl = (async () =>
      new Response(JSON.stringify({ access_token: "a2" }), { status: 200 })) as typeof fetch;
    const auth = new AuthClient({
      baseURL: "https://api.test",
      secure,
      access,
      fetchImpl,
    });
    await expect(auth.refreshOnce()).resolves.toBe(true);
    expect(access.get()).toBe("a2");
  });
});
