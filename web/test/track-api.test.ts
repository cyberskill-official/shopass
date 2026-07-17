import { apiFetch } from "../lib/api";
import { listTrackedProducts, trackShopeeProduct } from "../lib/track/api";

jest.mock("../lib/api", () => ({
  apiFetch: jest.fn(),
}));

describe("Track API", () => {
  const apiFetchMock = apiFetch as jest.MockedFunction<typeof apiFetch>;

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("submits a Shopee URL through the authenticated API client", async () => {
    apiFetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({ product_id: 42, platform: "shopee", already_tracked: false }),
    } as Response);

    await expect(trackShopeeProduct("https://shopee.vn/x-i.88123.20114455667")).resolves.toMatchObject({
      product_id: 42,
      platform: "shopee",
    });
    expect(apiFetchMock).toHaveBeenCalledWith("/v1/track", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        platform: "shopee",
        item_url: "https://shopee.vn/x-i.88123.20114455667",
      }),
    });
  });

  it("loads only the tracked-products collection endpoint", async () => {
    apiFetchMock.mockResolvedValue({
      ok: true,
      json: async () => [{
        product_id: 42,
        platform: "shopee",
        platform_item_id: "20114455667:88123",
        first_seen: "2026-07-17T10:00:00Z",
        tracked_at: "2026-07-17T10:00:00Z",
      }],
    } as Response);

    await expect(listTrackedProducts()).resolves.toHaveLength(1);
    expect(apiFetchMock).toHaveBeenCalledWith("/v1/tracked-products");
  });

  it("surfaces the server's safe validation message for an invalid link", async () => {
    apiFetchMock.mockResolvedValue({
      ok: false,
      json: async () => ({ error: "invalid item_url" }),
    } as Response);

    await expect(trackShopeeProduct("https://example.com/not-shopee")).rejects.toThrow("invalid item_url");
  });
});
