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

    expect(screen.getByRole("link", { name: /Shopass.*Bảng điều khiển/i })).toBeInTheDocument();
    expect(screen.getByText("Shop", { exact: true })).toBeInTheDocument();
    expect(screen.getByText("ass", { exact: true })).toBeInTheDocument();
    expect(screen.queryByText(/SănDeal/i)).not.toBeInTheDocument();
  });
});
