/**
 * @jest-environment jsdom
 */
import { render, screen } from "@testing-library/react";
import { LandingPage } from "../components/landing/landing-page";
import { LANDING_FAQ } from "../lib/landing/jsonld";

jest.mock("../components/landing/demo-chart", () => ({
  LandingDemoChart: () => <div data-testid="demo-chart">chart</div>,
}));

describe("R38 landing page", () => {
  it("shows Shopass hero, chart, trust links, and FAQ", () => {
    render(<LandingPage />);
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(/chạm đáy/i);
    expect(screen.getByTestId("demo-chart")).toBeTruthy();
    expect(screen.getByRole("link", { name: /cookie-stuffing/i })).toBeTruthy();
    expect(screen.getAllByRole("link", { name: /chính sách bảo mật/i }).length).toBeGreaterThan(0);
    for (const item of LANDING_FAQ) {
      expect(screen.getByText(item.q)).toBeTruthy();
    }
  });
});
