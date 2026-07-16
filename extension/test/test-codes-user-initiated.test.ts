// extension/test/test-codes-user-initiated.test.ts
import { testCodes } from "../src/content/shared/test-codes";
import * as fs from "fs";
import * as path from "path";

describe("TASK-CART-005 User Initiated constraint", () => {
  it("KHÔNG chạy khi không phải user-initiated", async () => {
    await expect(testCodes(["A", "B"], { userInitiated: false, cancelled: () => false }))
      .rejects.toThrow(/user chủ động/);
  });

  it("mã nguồn KHÔNG có setInterval/alarms gọi testCodes", async () => {
    const files = ["src/content/shared/test-codes.ts", "src/ui/test-codes-button.ts"];
    for (const f of files) {
      const src = await fs.promises.readFile(path.join(__dirname, "..", f), "utf8");
      expect(src).not.toMatch(/setInterval/);
      expect(src).not.toMatch(/chrome\.alarms/);
    }
  });
});
