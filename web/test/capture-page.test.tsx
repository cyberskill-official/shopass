import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import BrowserCapturePage from "../app/(app)/capture/page";
import { submitBrowserPrice } from "../lib/chart/fetch-chart";
import { trackShopeeProduct } from "../lib/track/api";

const mockReplace = jest.fn();
let mockSearchParams = new URLSearchParams();

jest.mock("next/navigation", () => ({
  useRouter: () => ({ replace: mockReplace }),
  useSearchParams: () => mockSearchParams,
}));

jest.mock("../lib/track/api", () => ({
  trackShopeeProduct: jest.fn(),
}));

jest.mock("../lib/chart/fetch-chart", () => ({
  submitBrowserPrice: jest.fn(),
}));

describe("Browser-assisted capture confirmation", () => {
  const trackShopeeProductMock = trackShopeeProduct as jest.MockedFunction<typeof trackShopeeProduct>;
  const submitBrowserPriceMock = submitBrowserPrice as jest.MockedFunction<typeof submitBrowserPrice>;

  beforeEach(() => {
    jest.clearAllMocks();
    mockSearchParams = new URLSearchParams({
      url: "https://www.shopee.vn/Ao-x-i.28863989.57657079476?sp_atk=discard-me",
      price: "6.490.000đ",
    });
  });

  it("requires customer confirmation, then tracks and saves the price", async () => {
    trackShopeeProductMock.mockResolvedValue({
      product_id: 77,
      platform: "shopee",
      already_tracked: false,
    });
    submitBrowserPriceMock.mockResolvedValue({ written: true });

    render(<BrowserCapturePage />);

    const itemURL = await screen.findByLabelText(/Liên kết sản phẩm Shopee/i);
    expect(itemURL).toHaveValue("https://shopee.vn/Ao-x-i.28863989.57657079476");
    expect(screen.getByLabelText(/Giá bạn đang thấy/i)).toHaveValue("6.490.000");

    fireEvent.click(screen.getByRole("button", { name: /Xác nhận và lưu vào biểu đồ/i }));

    await waitFor(() => {
      expect(trackShopeeProductMock).toHaveBeenCalledWith("https://shopee.vn/Ao-x-i.28863989.57657079476");
      expect(submitBrowserPriceMock).toHaveBeenCalledWith(77, 6_490_000);
      expect(mockReplace).toHaveBeenCalledWith("/products/77/chart?captured=1");
    });
  });

  it("rejects an invalid product URL before any backend write", async () => {
    render(<BrowserCapturePage />);
    const itemURL = await screen.findByLabelText(/Liên kết sản phẩm Shopee/i);
    fireEvent.change(itemURL, { target: { value: "https://example.com/not-a-shopee-product" } });
    fireEvent.change(screen.getByLabelText(/Giá bạn đang thấy/i), { target: { value: "6490000" } });
    fireEvent.click(screen.getByRole("button", { name: /Xác nhận và lưu vào biểu đồ/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(/Dán liên kết trực tiếp/i);
    expect(trackShopeeProductMock).not.toHaveBeenCalled();
    expect(submitBrowserPriceMock).not.toHaveBeenCalled();
  });

  it("allows manual price entry when a bookmarklet could not find one", async () => {
    mockSearchParams = new URLSearchParams({
      url: "https://shopee.vn/Ao-x-i.28863989.57657079476",
    });
    trackShopeeProductMock.mockResolvedValue({
      product_id: 91,
      platform: "shopee",
      already_tracked: true,
    });
    submitBrowserPriceMock.mockResolvedValue({ written: true });

    render(<BrowserCapturePage />);
    const price = await screen.findByLabelText(/Giá bạn đang thấy/i);
    expect(price).toHaveValue("");
    fireEvent.change(price, { target: { value: "199.000" } });
    fireEvent.click(screen.getByRole("button", { name: /Xác nhận và lưu vào biểu đồ/i }));

    await waitFor(() => {
      expect(submitBrowserPriceMock).toHaveBeenCalledWith(91, 199_000);
    });
  });
});
