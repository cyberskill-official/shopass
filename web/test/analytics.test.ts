import { __analyticsBuffer, __resetAnalyticsBuffer, trackEvent } from "../lib/analytics";

describe("analytics stub (R40 pending)", () => {
  beforeEach(() => {
    __resetAnalyticsBuffer();
  });

  it("buffers funnel events", () => {
    trackEvent("signup-click", { surface: "landing" });
    trackEvent("install-click");
    expect(__analyticsBuffer()).toEqual([
      { name: "signup-click", payload: { surface: "landing" } },
      { name: "install-click", payload: undefined },
    ]);
  });
});
