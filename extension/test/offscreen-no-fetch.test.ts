import * as fs from "fs";
import * as path from "path";

describe("offscreen KHÔNG tự fetch trang sàn (DEC-EXT-22)", () => {
  const offscreenDir = path.join(__dirname, "..", "src", "offscreen");

  test("offscreen.ts không có fetch tới domain sàn", () => {
    const src = fs.readFileSync(path.join(offscreenDir, "offscreen.ts"), "utf8");
    expect(src).not.toMatch(/fetch\(["'`]https:\/\/(shopee|tiktok|lazada)/);
  });

  test("manager.ts không có fetch tới domain sàn", () => {
    const src = fs.readFileSync(path.join(offscreenDir, "manager.ts"), "utf8");
    expect(src).not.toMatch(/fetch\(["'`]https:\/\/(shopee|tiktok|lazada)/);
  });

  test("offscreen.ts không import fetch/axios/got", () => {
    const src = fs.readFileSync(path.join(offscreenDir, "offscreen.ts"), "utf8");
    expect(src).not.toMatch(/import.*from\s+["'](node-fetch|axios|got)/);
  });
});
