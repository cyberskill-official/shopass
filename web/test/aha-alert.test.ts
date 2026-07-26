import { createAhaAlert } from "../lib/onboarding/aha-alert";
import { apiFetch } from "../lib/api";

jest.mock("../lib/api", () => ({
  apiFetch: jest.fn(),
}));

describe("createAhaAlert", () => {
  const apiFetchMock = apiFetch as jest.MockedFunction<typeof apiFetch>;

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("uses bottom_predicted when allowed", async () => {
    apiFetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({
        id: 1,
        product_id: 9,
        rule_type: "bottom_predicted",
        threshold: null,
        channel: ["push"],
        active: true,
      }),
    } as Response);

    await expect(createAhaAlert(9)).resolves.toMatchObject({
      rule_type: "bottom_predicted",
      premium_deferred: false,
    });
    expect(apiFetchMock).toHaveBeenCalledTimes(1);
  });

  it("falls back to real_sale on 402 premium gate", async () => {
    apiFetchMock
      .mockResolvedValueOnce({
        ok: false,
        status: 402,
        json: async () => ({ error: "premium_required" }),
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          id: 2,
          product_id: 9,
          rule_type: "real_sale",
          threshold: null,
          channel: ["push"],
          active: true,
        }),
      } as Response);

    await expect(createAhaAlert(9)).resolves.toMatchObject({
      rule_type: "real_sale",
      premium_deferred: true,
    });
    expect(apiFetchMock).toHaveBeenCalledTimes(2);
  });
});
