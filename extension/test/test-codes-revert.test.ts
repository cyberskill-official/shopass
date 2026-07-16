// extension/test/test-codes-revert.test.ts
import { testCodes } from "../src/content/shared/test-codes";
import * as applier from "../src/content/shared/code-applier";
import * as fs from "fs";
import * as path from "path";
import * as pacing from "../src/content/shared/pacing";

describe("TASK-CART-005 Revert constraint", () => {
  beforeEach(() => {
    jest.spyOn(pacing, "sleep").mockResolvedValue(); // don't sleep in revert tests
  });
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("revert sau mỗi mã; KHÔNG chốt đơn", async () => {
    jest.spyOn(applier, "userInitiatedApply").mockResolvedValue({ valid: true, discount: 30_000 });
    const revertSpy = jest.spyOn(applier, "revert").mockImplementation(() => {});
    await testCodes(["A", "B", "C"], { userInitiated: true, cancelled: () => false });
    expect(revertSpy).toHaveBeenCalledTimes(3); // mỗi mã gỡ một lần
  });

  it("mã nguồn KHÔNG có bước chốt đơn", async () => {
    const src = await fs.promises.readFile(path.join(__dirname, "../src/content/shared/code-applier.ts"), "utf8");
    expect(src).not.toMatch(/place.?order|checkout|chốt đơn|submitOrder/i);
  });

  it("sortDesc theo discount", async () => {
    const seq = [{ valid: true, discount: 20_000 }, { valid: true, discount: 50_000 }, { valid: true, discount: 30_000 }];
    let i = 0;
    jest.spyOn(applier, "userInitiatedApply").mockImplementation(async () => seq[i++]);
    jest.spyOn(applier, "revert").mockImplementation(() => {});
    const res = await testCodes(["A", "B", "C"], { userInitiated: true, cancelled: () => false });
    expect(res.map((r) => r.discount)).toEqual([50_000, 30_000, 20_000]); // giảm nhiều nhất đầu
  });
});
