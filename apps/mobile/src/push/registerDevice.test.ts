import { bootstrapPush, registerDevice, unregisterDevice, watchTokenRefresh } from "./registerDevice";
import { HttpClient } from "../api/httpClient";
import { AuthClient } from "../auth/authClient";
import { MemoryAccessToken, MemorySecureStore } from "../auth/tokenStore";

describe("registerDevice", () => {
  it("POSTs device token and unregisters on logout path", async () => {
    const methods: string[] = [];
    const fetchImpl = (async (_u: RequestInfo | URL, init?: RequestInit) => {
      methods.push(init?.method ?? "GET");
      return new Response(null, { status: init?.method === "DELETE" ? 204 : 200 });
    }) as typeof fetch;
    const access = new MemoryAccessToken();
    access.set("t");
    const auth = new AuthClient({
      baseURL: "https://api.test",
      secure: new MemorySecureStore(),
      access,
      fetchImpl,
    });
    const http = new HttpClient("https://api.test", access, auth, fetchImpl);
    await registerDevice(http, "fcm-1", "ios");
    await unregisterDevice(http, "fcm-1");
    expect(methods).toEqual(["POST", "DELETE"]);
  });

  it("skips register when permission denied", async () => {
    const access = new MemoryAccessToken();
    access.set("t");
    const fetchImpl = (async () => {
      throw new Error("should not call");
    }) as typeof fetch;
    const auth = new AuthClient({
      baseURL: "https://api.test",
      secure: new MemorySecureStore(),
      access,
      fetchImpl,
    });
    const http = new HttpClient("https://api.test", access, auth, fetchImpl);
    const token = await bootstrapPush(
      http,
      {
        requestPermission: async () => false,
        getToken: async () => "x",
        onTokenRefresh: () => () => undefined,
      },
      "android",
    );
    expect(token).toBeNull();
  });

  it("refreshes token via onTokenRefresh", async () => {
    let cb: ((t: string) => void) | undefined;
    const posts: string[] = [];
    const fetchImpl = (async (_u: RequestInfo | URL, init?: RequestInit) => {
      if (init?.method === "POST") posts.push(String(init.body));
      return new Response(null, { status: 200 });
    }) as typeof fetch;
    const access = new MemoryAccessToken();
    access.set("t");
    const auth = new AuthClient({
      baseURL: "https://api.test",
      secure: new MemorySecureStore(),
      access,
      fetchImpl,
    });
    const http = new HttpClient("https://api.test", access, auth, fetchImpl);
    watchTokenRefresh(
      http,
      {
        requestPermission: async () => true,
        getToken: async () => "a",
        onTokenRefresh: (fn) => {
          cb = fn;
          return () => undefined;
        },
      },
      "ios",
    );
    cb?.("rotated");
    await new Promise((r) => setTimeout(r, 0));
    expect(posts.some((b) => b.includes("rotated"))).toBe(true);
  });
});
