import React from "react";
import { render, screen } from "@testing-library/react";
import { PriceChart } from "../components/price-chart/price-chart";

// Mock recharts to avoid jsdom issues with SVG measurement
jest.mock('recharts', () => {
  const Original = jest.requireActual('recharts');
  return {
    ...Original,
    ResponsiveContainer: ({ children }: any) => <div>{children}</div>,
  };
});

const feed = {
  product_id: 90112,
  range: "90d",
  maturity: "MATURE",
  daily: [
    { day: "2026-04-04", min_p: 119000, max_p: 149000, close_p: 119000 },
    { day: "2026-05-05", min_p: 109000, max_p: 139000, close_p: 109000 },
  ],
  annotations: { 
    median90: 129000, 
    trailing_min: 99000, 
    verdict: "TAM_DUOC",
    accumulating: false, 
    double_dates: ["2026-04-04", "2026-05-05"] 
  },
};

describe("PriceChart", () => {
  it("vẽ đường tham chiếu median90 + trailing_min từ feed (không tự tính)", () => {
    render(<PriceChart data={feed as any} />);
    expect(screen.getByTestId("ref-median90")).toHaveAttribute("data-value", "129000");
    expect(screen.getByTestId("ref-trailing-min")).toHaveAttribute("data-value", "99000");
  });

  it("đánh dấu mốc ngày đôi từ feed", () => {
    render(<PriceChart data={feed as any} />);
    expect(screen.getAllByTestId("double-date-marker").length).toBe(2);
  });
  
  it("hiển thị trạng thái đang thu thập dữ liệu khi daily rỗng", () => {
    render(<PriceChart data={{ ...feed, daily: [] } as any} />);
    expect(screen.getByText(/Đang thu thập dữ liệu giá/i)).toBeInTheDocument();
  });
});
