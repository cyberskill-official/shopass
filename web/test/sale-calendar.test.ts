import { daysUntil, nextDoubleDate } from "../lib/tools/sale-calendar";

describe("sale calendar helpers", () => {
  it("picks the next double date on or after from", () => {
    const { date, label } = nextDoubleDate(new Date("2026-07-26T12:00:00Z"));
    expect(label).toBe("8.8");
    expect(date.toISOString().slice(0, 10)).toBe("2026-08-08");
  });

  it("counts whole UTC days until target", () => {
    expect(daysUntil(new Date("2026-08-08T00:00:00Z"), new Date("2026-07-26T15:00:00Z"))).toBe(13);
  });
});
