import { displayVerdict, getPriceHistory, listWishlist } from "./trackClient";
import { HttpClient } from "../api/httpClient";
import { AuthClient } from "../auth/authClient";
import { MemoryAccessToken, MemorySecureStore } from "../auth/tokenStore";

function fakeFetch(handler: (url: string) => Response): typeof fetch {
  return (async (input: RequestInfo | URL) => handler(String(input))) as typeof fetch;
}

describe("trackClient", () => {
  it("reads wishlist from API (thin client)", async () => {
    const fetchImpl = fakeFetch((url) => {
      if (url.includes("/v1/wishlists")) {
        return new Response(JSON.stringify([{ product_id: 9, verdict: "UNKNOWN" }]), {
          status: 200,
        });
      }
      throw new Error(url);
    });
    const access = new MemoryAccessToken();
    access.set("tok");
    const auth = new AuthClient({
      baseURL: "https://api.test",
      secure: new MemorySecureStore(),
      access,
      fetchImpl,
    });
    const http = new HttpClient("https://api.test", access, auth, fetchImpl);
    const items = await listWishlist(http);
    expect(items[0].product_id).toBe(9);
    expect(displayVerdict(items[0].verdict)).toBe("chưa đủ dữ liệu");
  });

  it("loads price-history without client-side sale math", async () => {
    const fetchImpl = fakeFetch((url) => {
      expect(url).toContain("/price-history");
      return new Response(JSON.stringify({ points: [{ day: "2026-06-20", close_p: 100 }] }), {
        status: 200,
      });
    });
    const access = new MemoryAccessToken();
    access.set("tok");
    const auth = new AuthClient({
      baseURL: "https://api.test",
      secure: new MemorySecureStore(),
      access,
      fetchImpl,
    });
    const http = new HttpClient("https://api.test", access, auth, fetchImpl);
    const pts = await getPriceHistory(http, 42);
    expect(pts).toHaveLength(1);
  });
});
