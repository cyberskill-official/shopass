import type { Page, Request } from "playwright";

export interface Outbound {
  url: string;
  headers: Record<string, string>;
  body: string;
}

const BACKEND_HOST = "api.sandeal.vn"; // endpoint hợp lệ DUY NHẤT
const CRED_RE = /(SPC_|session|sessionid|token|bearer|password|mật khẩu|@[\w.]+\.\w+)/i;

export function installTrap(page: Page, captured: Outbound[]) {
  page.on("request", (req: Request) => {
    captured.push({
      url: req.url(),
      headers: req.headers(),
      body: req.postData() ?? "",
    });
  });
}

export function assertNoCredentialEgress(captured: Outbound[]) {
  for (const o of captured) {
    let host;
    try {
      host = new URL(o.url).host;
    } catch {
      continue; // bỏ qua data URI hoặc invalid
    }
    
    if (host !== BACKEND_HOST) {
      throw new Error(`outbound tới host lạ: ${host} (chỉ cho ${BACKEND_HOST})`);
    }

    const blob = `${o.url}\n${JSON.stringify(o.headers)}\n${o.body}`;
    if (CRED_RE.test(blob)) {
      throw new Error(`PHÁT HIỆN credential/PII rời máy trong request tới ${host}`);
    }
  }
}
