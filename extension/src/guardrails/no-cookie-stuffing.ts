import { SourceFile, Violation } from './types';

const SHOP_HOSTS = ["shopee.vn", "tiktok.com", "lazada.vn"];

// Pattern bị cấm: set cookie / chrome.cookies.set / điều hướng affiliate trên host sàn.
const BANNED = [
  /chrome\.cookies\.set/,
  /document\.cookie\s*=/,
  /chrome\.tabs\.update\([^)]*affiliate/i,
  /window\.open\([^)]*aff/i,        // pop-under affiliate
];

function touchesShopHost(f: SourceFile, ln: string, hosts: string[]): boolean {
    const lowerLine = ln.toLowerCase();
    const lowerPath = f.path.toLowerCase();
    for (const host of hosts) {
        if (lowerLine.includes(host) || lowerPath.includes(host.split('.')[0])) {
            return true;
        }
    }
    if (lowerPath.includes("content") && (ln.includes("document.cookie") || ln.includes("chrome.cookies.set"))) {
        return true; 
    }
    // Any line trying to open an affiliate link or redirect to affiliate should be flagged
    if (lowerLine.includes("window.open") && lowerLine.includes("aff")) {
        return true;
    }
    if (lowerLine.includes("chrome.tabs.update") && lowerLine.includes("affiliate")) {
        return true;
    }
    return false;
}

// scanSource trả danh sách vi phạm {file, line, rule, policyRef}. Rỗng = pass.
export function scanSource(files: SourceFile[]): Violation[] {
  const out: Violation[] = [];
  for (const f of files) {
    // skip test files
    if (f.path.includes('.test.ts') || f.path.includes('test/')) continue;
    f.lines.forEach((ln, i) => {
      for (const re of BANNED) {
        if (re.test(ln) && touchesShopHost(f, ln, SHOP_HOSTS)) {
          out.push({ file: f.path, line: i + 1, rule: re.source,
            policyRef: "Chrome Web Store affiliate policy 2025-03 (enforced 2025-06-10): no cookie/redirect injection without user benefit + action" });
        }
      }
    });
  }
  return out;
}
