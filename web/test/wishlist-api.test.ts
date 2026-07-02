import { addItem } from "../lib/wishlist/api";
import { apiFetch } from "../lib/api";

jest.mock("../lib/api", () => ({
  apiFetch: jest.fn(),
}));

describe("Wishlist API", () => {
  let apiFetchMock: jest.Mock;

  beforeEach(() => {
    apiFetchMock = apiFetch as jest.Mock;
    apiFetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({ id: 1, productId: 90112, targetPrice: null }),
    });
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it("addItem gửi target_price là số nguyên VND (không float)", async () => {
    await addItem(3, 90112, 99000);
    expect(apiFetchMock).toHaveBeenCalled();
    const callArgs = apiFetchMock.mock.calls[0][1];
    const body = JSON.parse(callArgs.body);
    
    expect(body).toEqual({ product_id: 90112, target_price: 99000 });
    expect(Number.isInteger(body.target_price)).toBe(true);
  });

  it("addItem để trống target_price gửi null", async () => {
    await addItem(3, 90112, null);
    const callArgs = apiFetchMock.mock.calls[0][1];
    const body = JSON.parse(callArgs.body);
    expect(body.target_price).toBeNull();
  });
});
