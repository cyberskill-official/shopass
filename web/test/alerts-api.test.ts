import { createAlert, listAlerts } from "../lib/alerts/api";
import { PremiumRequiredError } from "../lib/alerts/errors";
import { apiFetch } from "../lib/api";

jest.mock("../lib/api", () => ({
  apiFetch: jest.fn(),
}));

describe("alerts api", () => {
  const apiFetchMock = apiFetch as jest.MockedFunction<typeof apiFetch>;

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("throws PremiumRequiredError on 402", async () => {
    apiFetchMock.mockResolvedValue({
      ok: false,
      status: 402,
      json: async () => ({ error: "premium_required" }),
    } as Response);

    await expect(
      createAlert({
        product_id: 1,
        rule_type: "bottom_predicted",
        threshold: null,
        channels: ["push"],
      }),
    ).rejects.toBeInstanceOf(PremiumRequiredError);
  });

  it("maps channel array from list response", async () => {
    apiFetchMock.mockResolvedValue({
      ok: true,
      json: async () => [
        {
          id: 1,
          product_id: 9,
          rule_type: "real_sale",
          threshold: null,
          channel: ["push"],
          active: true,
        },
      ],
    } as Response);

    await expect(listAlerts()).resolves.toEqual([
      {
        id: 1,
        product_id: 9,
        rule_type: "real_sale",
        threshold: null,
        channels: ["push"],
        active: true,
      },
    ]);
  });
});
