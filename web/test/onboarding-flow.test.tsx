/**
 * @jest-environment jsdom
 */
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { OnboardingFlow } from "../components/onboarding/onboarding-flow";
import { __analyticsBuffer, __resetAnalyticsBuffer } from "../lib/analytics";
import { fetchChart } from "../lib/chart/fetch-chart";
import { createAhaAlert } from "../lib/onboarding/aha-alert";
import { trackShopeeProduct } from "../lib/track/api";

jest.mock("../lib/track/api", () => ({
  trackShopeeProduct: jest.fn(),
}));
jest.mock("../lib/chart/fetch-chart", () => ({
  fetchChart: jest.fn(),
}));
jest.mock("../lib/onboarding/aha-alert", () => ({
  createAhaAlert: jest.fn(),
}));
jest.mock("../components/price-chart/price-chart", () => ({
  PriceChart: () => <div data-testid="price-chart">chart</div>,
}));

describe("R45 onboarding flow", () => {
  beforeEach(() => {
    __resetAnalyticsBuffer();
    jest.clearAllMocks();
  });

  it("tracks URL, shows chart step, then creates aha alert", async () => {
    (trackShopeeProduct as jest.Mock).mockResolvedValue({
      product_id: 42,
      platform: "shopee",
      already_tracked: false,
    });
    (fetchChart as jest.Mock).mockResolvedValue({
      product_id: 42,
      range: "90d",
      maturity: "MATURE",
      daily: [{ day: "2026-07-01", min_p: 1, max_p: 2, close_p: 1 }],
      annotations: {
        median90: 1,
        trailing_min: 1,
        verdict: "TAM_DUOC",
        accumulating: false,
        double_dates: [],
      },
    });
    (createAhaAlert as jest.Mock).mockResolvedValue({
      rule: { id: 1 },
      rule_type: "real_sale",
      premium_deferred: true,
    });

    render(<OnboardingFlow />);
    fireEvent.change(screen.getByLabelText(/link sản phẩm/i), {
      target: { value: "https://shopee.vn/x-i.1.2" },
    });
    fireEvent.click(screen.getByRole("button", { name: /theo dõi & xem biểu đồ/i }));

    await waitFor(() => {
      expect(screen.getByText(/sản phẩm #42/i)).toBeTruthy();
    });
    expect(__analyticsBuffer().some((e) => e.name === "first-track")).toBe(true);

    fireEvent.click(screen.getByRole("button", { name: /báo tôi khi chạm đáy/i }));
    await waitFor(() => {
      expect(screen.getByText(/cảnh báo đã bật/i)).toBeTruthy();
    });
    expect(__analyticsBuffer().some((e) => e.name === "first-alert")).toBe(true);
  });
});
