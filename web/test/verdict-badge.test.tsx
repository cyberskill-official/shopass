import React from "react";
import { render, screen } from "@testing-library/react";
import { VerdictBadge } from "../components/price-chart/verdict-badge";

describe("VerdictBadge", () => {
  it("badge lấy thẳng verdict từ feed", () => {
    render(<VerdictBadge verdict="SALE_AO" maturity="MATURE" />);
    expect(screen.getByText("Sale ảo")).toHaveAttribute("data-verdict", "SALE_AO");
  });

  it("SKU NEW (<14 ngày) KHÔNG hiện badge kết luận", () => {
    const { container } = render(<VerdictBadge verdict="UNKNOWN" maturity="NEW" />);
    expect(container).toBeEmptyDOMElement();
  });
});
