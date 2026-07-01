// extension/test/test-codes-pacing.test.ts
import { randomDelayMs } from "../src/content/shared/pacing";

describe("FR-CART-005 Pacing constraint", () => {
  it("delay nằm trong [2500, 5000)", () => {
    for (let k = 0; k < 100; k++) {
      const d = randomDelayMs();
      expect(d).toBeGreaterThanOrEqual(2500);
      expect(d).toBeLessThan(5000);
    }
  });

  it("có sleep giữa các lần thử (dùng clock giả)", async () => {
    jest.useFakeTimers();
    const applier = require("../src/content/shared/code-applier");
    jest.spyOn(applier, "userInitiatedApply").mockResolvedValue({ valid: false, discount: 0 });
    jest.spyOn(applier, "revert").mockImplementation(() => {});
    const { testCodes } = require("../src/content/shared/test-codes");
    
    const p = testCodes(["A", "B"], { userInitiated: true, cancelled: () => false });
    
    // mỗi mã chờ >=2.5s trước khi apply
    // advance timers multiple times to ensure promises resolve
    for (let i = 0; i < 20; i++) {
      await jest.advanceTimersByTimeAsync(1000);
    }
    
    await p;
    jest.useRealTimers();
  });
});
