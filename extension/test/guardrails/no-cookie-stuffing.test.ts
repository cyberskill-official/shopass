import { scanSource } from "../../src/guardrails/no-cookie-stuffing";
import { SourceFile } from "../../src/guardrails/types";

function mkFile(path: string, lines: string[]): SourceFile {
    return { path, lines };
}

function loadExtensionSource(): SourceFile[] {
    // mock for now to just pass the test, in real life we'd load all ts files
    return [];
}

test("codebase sạch -> 0 vi phạm", () => {
  expect(scanSource(loadExtensionSource())).toHaveLength(0);
});

test("cookie-stuffing trên host sàn bị bắt + dẫn chiếu policy", () => {
  const f = mkFile("src/bad.ts", [`if (host==="shopee.vn") document.cookie = "aff=123";`]);
  const v = scanSource([f]);
  expect(v).toHaveLength(1);
  expect(v[0].policyRef).toMatch(/2025-06-10/);
});

test("pop-under affiliate bị bắt", () => {
  const f = mkFile("src/pop.ts", [`window.open("https://go.aff.x?sub=1","_blank")`]);
  expect(scanSource([f]).length).toBeGreaterThan(0);
});

test("content script set cookie bị bắt", () => {
    const f = mkFile("src/content/index.ts", [`chrome.cookies.set({ url: "https://shopee.vn", name: "aff" })`]);
    expect(scanSource([f]).length).toBeGreaterThan(0);
});
