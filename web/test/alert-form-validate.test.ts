import { validateAlert } from "../lib/alerts/validate";

describe("validateAlert", () => {
  it("price_below cần số nguyên dương VND", () => {
    expect(validateAlert("price_below", 0, ["push"])).toMatch(/số nguyên dương/i);
    expect(validateAlert("price_below", 89000, ["push"])).toBeNull();
  });

  it("drop_pct cần 1..99", () => {
    expect(validateAlert("drop_pct", 0, ["push"])).toMatch(/1\.\.99/);
    expect(validateAlert("drop_pct", 150, ["push"])).toMatch(/1\.\.99/);
    expect(validateAlert("drop_pct", 30, ["push"])).toBeNull();
  });

  it("real_sale / bottom_predicted KHÔNG nhận threshold", () => {
    expect(validateAlert("real_sale", 100, ["push"])).toMatch(/không nhận ngưỡng/i);
    expect(validateAlert("bottom_predicted", null, ["push"])).toBeNull();
  });
});
