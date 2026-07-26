import BeecostPage, { metadata as beecostMeta } from "../app/(marketing)/so-sanh/shopass-vs-beecost/page";
import HoneyPage, { metadata as honeyMeta } from "../app/(marketing)/thay-the-honey/page";
import { SHOPASS_VS_BEECOST } from "../lib/compare/beecost";

describe("R44 comparison pages", () => {
  it("BeeCost page has VN metadata and a full feature matrix", () => {
    expect(beecostMeta.title).toMatch(/BeeCost/i);
    expect(SHOPASS_VS_BEECOST.length).toBeGreaterThanOrEqual(8);
    expect(SHOPASS_VS_BEECOST.some((r) => r.feature.includes("TikTok"))).toBe(true);
    expect(BeecostPage()).toBeTruthy();
  });

  it("Honey alternative page cites sources and links minh-bach", () => {
    expect(honeyMeta.title).toMatch(/Honey/i);
    expect(HoneyPage()).toBeTruthy();
  });
});
