/**
 * @jest-environment jsdom
 */
import { render } from "@testing-library/react";
import { axe, toHaveNoViolations } from "jest-axe";
import { LandingPage } from "../components/landing/landing-page";

expect.extend(toHaveNoViolations);

jest.mock("../components/landing/demo-chart", () => ({
  LandingDemoChart: () => (
    <div role="img" aria-label="Biểu đồ giá minh họa">
      chart
    </div>
  ),
}));

describe("R48 a11y — landing", () => {
  it("has no serious axe violations", async () => {
    const { container } = render(<LandingPage />);
    const results = await axe(container, {
      rules: {
        // json-ld scripts are not user-facing; axe flags region/landmark noise sometimes
        "region": { enabled: false },
      },
    });
    expect(results).toHaveNoViolations();
  });
});
