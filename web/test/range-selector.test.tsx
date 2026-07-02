import { fetchChart } from "../lib/chart/fetch-chart";
import { RANGE_ALLOWLIST } from "../lib/chart/types";
import { apiFetch } from "../lib/api";

jest.mock("../lib/api", () => ({
  apiFetch: jest.fn(),
}));

describe("Range Selector (fetchChart validation)", () => {
  it("chỉ allowlist range; range lạ ném lỗi, không gọi mạng", async () => {
    const apiFetchMock = apiFetch as jest.Mock;
    apiFetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({}),
    });

    await expect(fetchChart(1, "5d" as any)).rejects.toThrow(/allowlist/);
    expect(apiFetchMock).not.toHaveBeenCalled();
    expect(RANGE_ALLOWLIST).toEqual(["7d", "30d", "90d", "180d", "1y"]);
  });
});
