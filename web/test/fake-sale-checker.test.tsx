/**
 * @jest-environment jsdom
 */
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { FakeSaleChecker } from "../components/tools/fake-sale-checker";
import { __analyticsBuffer, __resetAnalyticsBuffer } from "../lib/analytics";

describe("R43 fake-sale checker", () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    __resetAnalyticsBuffer();
  });

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it("shows verdict for a tracked product", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        tracked: true,
        platform: "shopee",
        product_id: 1,
        maturity: "MATURE",
        verdict: "SALE_AO",
        current_price: 100000,
        median90: 150000,
        trailing_min: 90000,
        daily: [],
      }),
    } as Response);

    render(<FakeSaleChecker />);
    fireEvent.change(screen.getByLabelText(/link sản phẩm/i), {
      target: { value: "https://shopee.vn/x-i.1.2" },
    });
    fireEvent.click(screen.getByRole("button", { name: /^kiểm tra$/i }));

    await waitFor(() => {
      expect(screen.getByText(/sale ảo/i)).toBeTruthy();
    });
    expect(__analyticsBuffer().some((e) => e.name === "verdict-shown")).toBe(true);
  });

  it("captures lead when untracked", async () => {
    global.fetch = jest
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ tracked: false, platform: "shopee", message: "not_tracked" }),
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ ok: true, id: 11 }),
      } as Response);

    render(<FakeSaleChecker />);
    fireEvent.change(screen.getByLabelText(/link sản phẩm/i), {
      target: { value: "https://shopee.vn/x-i.1.2" },
    });
    fireEvent.click(screen.getByRole("button", { name: /^kiểm tra$/i }));

    await waitFor(() => {
      expect(screen.getByText(/chưa có trong shopass/i)).toBeTruthy();
    });

    fireEvent.change(screen.getByPlaceholderText(/you@email.com/i), {
      target: { value: "lead@example.com" },
    });
    fireEvent.click(screen.getByRole("button", { name: /báo tôi/i }));

    await waitFor(() => {
      expect(screen.getByText(/đã ghi nhận email/i)).toBeTruthy();
    });
    expect(__analyticsBuffer().some((e) => e.name === "lead-captured")).toBe(true);
  });
});
