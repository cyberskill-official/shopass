import * as fs from "fs";
import * as path from "path";

describe("sync-no-platform-token", () => {
  it("source code does not read document.cookie", () => {
    const files = ["auth-bridge.ts", "sender.ts", "ws-client.ts", "queue.ts"];
    for (const file of files) {
      const p = path.join(__dirname, "..", "src", "sync", file);
      if (fs.existsSync(p)) {
        const content = fs.readFileSync(p, "utf-8");
        expect(content).not.toMatch(/document\.cookie/);
        expect(content).not.toMatch(/cookie/); // broader check
      }
    }
  });
});
