import { render, screen } from "@testing-library/react";
import { AppShell } from "../components/app-shell";

jest.mock("next/navigation", () => ({
  usePathname: () => "/dashboard",
}));

jest.mock("../lib/auth", () => ({
  logout: jest.fn(),
}));

describe("App shell brand", () => {
  it("shows the Shopass product name in the signed-in experience", () => {
    render(<AppShell><p>Nội dung thử</p></AppShell>);

    const brandLink = screen.getByRole("link", { name: "Shopass — Bảng điều khiển" });
    expect(brandLink).toHaveAttribute("href", "/dashboard");
    expect(brandLink).toHaveTextContent("Shopass");
    expect(brandLink).not.toHaveTextContent(/Săn\s*Deal/i);
  });
});
