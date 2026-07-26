/**
 * @jest-environment jsdom
 */
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { PricingPage } from "../components/pricing/pricing-page";
import { __analyticsBuffer, __resetAnalyticsBuffer } from "../lib/analytics";

describe("R39 pricing page", () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    __resetAnalyticsBuffer();
  });

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it("shows Free / Premium / Pro prices and fires pricing-view", () => {
    render(<PricingPage />);
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(/bảng giá shopass/i);
    expect(screen.getByText("0đ")).toBeTruthy();
    expect(screen.getByText("29.000đ/tháng")).toBeTruthy();
    expect(screen.getByText("79.000đ/tháng")).toBeTruthy();
    expect(__analyticsBuffer().some((e) => e.name === "pricing-view")).toBe(true);
  });

  it("opens waitlist modal and submits lead", async () => {
    const fetchMock = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true, id: 7 }),
    } as Response);
    global.fetch = fetchMock;

    render(<PricingPage />);
    fireEvent.click(screen.getAllByRole("button", { name: /đăng ký chờ/i })[0]);
    expect(screen.getByRole("dialog")).toBeTruthy();
    expect(__analyticsBuffer().some((e) => e.name === "tier-click")).toBe(true);

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: "buyer@example.com" } });
    fireEvent.click(screen.getByRole("button", { name: /giữ chỗ/i }));

    await waitFor(() => {
      expect(screen.getByText(/đã ghi nhận/i)).toBeTruthy();
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/waitlist",
      expect.objectContaining({ method: "POST" }),
    );
    expect(__analyticsBuffer().some((e) => e.name === "waitlist-submit")).toBe(true);
  });
});
