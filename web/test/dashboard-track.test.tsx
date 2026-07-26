import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import DashboardPage from "../app/(app)/dashboard/page";
import { listTrackedProducts, trackShopeeProduct } from "../lib/track/api";

const mockPush = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: mockPush }),
}));

jest.mock("../lib/track/api", () => ({
  listTrackedProducts: jest.fn(),
  trackShopeeProduct: jest.fn(),
}));

jest.mock("../lib/referral/api", () => ({
  getReferralMe: jest.fn().mockResolvedValue({
    code: "SDTEST1",
    uses: 0,
    has_referrer: false,
    reward_note: "test",
  }),
}));

describe("Dashboard tracking flow", () => {
  const listTrackedProductsMock = listTrackedProducts as jest.MockedFunction<typeof listTrackedProducts>;
  const trackShopeeProductMock = trackShopeeProduct as jest.MockedFunction<typeof trackShopeeProduct>;

  beforeEach(() => {
    jest.clearAllMocks();
    listTrackedProductsMock.mockResolvedValue([]);
  });

  it("offers a Shopee URL form instead of fabricated deal data", async () => {
    render(<DashboardPage />);

    expect(await screen.findByText(/Chưa có sản phẩm nào được theo dõi/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Liên kết sản phẩm Shopee Việt Nam/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /bắt đầu: link → biểu đồ → cảnh báo/i })).toHaveAttribute(
      "href",
      "/onboarding",
    );
    expect(screen.queryByText(/Tai nghe Bluetooth SănDeal/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/Push & email/i)).not.toBeInTheDocument();
  });

  it("tracks a pasted URL and opens its chart", async () => {
    trackShopeeProductMock.mockResolvedValue({
      product_id: 77,
      platform: "shopee",
      already_tracked: false,
    });
    render(<DashboardPage />);

    await screen.findByText(/Chưa có sản phẩm nào được theo dõi/i);
    fireEvent.change(screen.getByLabelText(/Liên kết sản phẩm Shopee Việt Nam/i), {
      target: { value: "https://shopee.vn/x-i.88123.20114455667" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Theo dõi giá" }));

    await waitFor(() => {
      expect(trackShopeeProductMock).toHaveBeenCalledWith("https://shopee.vn/x-i.88123.20114455667");
      expect(mockPush).toHaveBeenCalledWith("/products/77/chart");
    });
  });

  it("lists only products returned by the owner-scoped endpoint", async () => {
    listTrackedProductsMock.mockResolvedValue([{
      product_id: 88,
      platform: "shopee",
      platform_item_id: "20114455667:88123",
      first_seen: "2026-07-17T10:00:00Z",
      tracked_at: "2026-07-17T10:00:00Z",
    }]);
    render(<DashboardPage />);

    expect(await screen.findByText(/Sản phẩm #88/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Xem biểu đồ" })).toHaveAttribute("href", "/products/88/chart");
  });
});
