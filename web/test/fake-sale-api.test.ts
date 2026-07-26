import { checkFakeSale } from "../lib/tools/fake-sale";

describe("fake-sale client", () => {
  const originalFetch = global.fetch;

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it("posts item_url to the BFF", async () => {
    const fetchMock = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        tracked: true,
        platform: "shopee",
        product_id: 9,
        maturity: "MATURE",
        verdict: "SALE_AO",
        current_price: 100000,
        median90: 120000,
        trailing_min: 90000,
        daily: [],
      }),
    } as Response);
    global.fetch = fetchMock;

    await expect(checkFakeSale("https://shopee.vn/x-i.1.2")).resolves.toMatchObject({
      tracked: true,
      verdict: "SALE_AO",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/tools/fake-sale-check", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ item_url: "https://shopee.vn/x-i.1.2" }),
    });
  });
});
