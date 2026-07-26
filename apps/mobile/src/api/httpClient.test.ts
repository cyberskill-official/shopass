import { HttpClient } from "./httpClient";
import { AuthClient } from "../auth/authClient";
import { MemoryAccessToken, MemorySecureStore } from "../auth/tokenStore";

describe("httpClient", () => {
  it("retries once after 401 via refresh", async () => {
    let calls = 0;
    const fetchImpl = (async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes("/v1/auth/refresh")) {
        return new Response(JSON.stringify({ access_token: "new" }), { status: 200 });
      }
      calls++;
      if (calls === 1) return new Response("nope", { status: 401 });
      return new Response(JSON.stringify({ ok: true }), { status: 200 });
    }) as typeof fetch;

    const access = new MemoryAccessToken();
    access.set("old");
    const secure = new MemorySecureStore();
    await secure.setRefresh("r");
    const auth = new AuthClient({
      baseURL: "https://api.test",
      secure,
      access,
      fetchImpl,
    });
    const http = new HttpClient("https://api.test", access, auth, fetchImpl);
    const res = await http.request("/v1/wishlists");
    expect(res.status).toBe(200);
    expect(access.get()).toBe("new");
  });
});
