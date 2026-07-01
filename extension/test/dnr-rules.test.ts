import * as fs from "fs";
import * as path from "path";

describe("DNR rules (DEC-EXT-20, DEC-EXT-21)", () => {
  test("DNR rule là tập tối thiểu (≤5 rules)", () => {
    const rulesPath = path.join(__dirname, "..", "src", "dnr", "rules.json");
    const rules = JSON.parse(fs.readFileSync(rulesPath, "utf8"));
    expect(Array.isArray(rules)).toBe(true);
    expect(rules.length).toBeLessThanOrEqual(5); // tối thiểu
  });

  test("dnr.ts KHÔNG dùng webRequest blocking (DEC-EXT-20)", () => {
    const dnrPath = path.join(__dirname, "..", "src", "dnr", "dnr.ts");
    const src = fs.readFileSync(dnrPath, "utf8");
    // Check only non-comment lines for webRequest usage
    const codeLines = src
      .split("\n")
      .filter((l) => !l.trim().startsWith("*") && !l.trim().startsWith("//"));
    const codeOnly = codeLines.join("\n");
    expect(codeOnly).not.toMatch(/chrome\.webRequest/);
    expect(codeOnly).not.toMatch(/onBeforeRequest/);
  });

  test("không có file nào trong src/ dùng webRequest blocking", () => {
    const srcDir = path.join(__dirname, "..", "src");
    const tsFiles = walkDir(srcDir).filter((f) => f.endsWith(".ts"));
    for (const file of tsFiles) {
      const content = fs.readFileSync(file, "utf8");
      expect(content).not.toMatch(/chrome\.webRequest.*blocking/);
    }
  });

  test("DNR rules.json không có rule rộng chặn/sửa request sàn (DEC-EXT-21)", () => {
    const rulesPath = path.join(__dirname, "..", "src", "dnr", "rules.json");
    const rules = JSON.parse(fs.readFileSync(rulesPath, "utf8")) as any[];
    for (const rule of rules) {
      // Không rule nào nhắm tới domain sàn với action block/redirect
      if (rule.condition?.urlFilter) {
        expect(rule.condition.urlFilter).not.toMatch(
          /shopee|tiktok|lazada/
        );
      }
    }
  });
});

function walkDir(dir: string): string[] {
  const result: string[] = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      result.push(...walkDir(full));
    } else {
      result.push(full);
    }
  }
  return result;
}
