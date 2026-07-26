import * as fs from "fs";
import * as path from "path";
import {
  ALL_HOST_PERMISSIONS,
  ALLOWED_API_HOSTNAMES,
  isAllowedApiUrl,
} from "../src/shared/allowlist";
import { API_ORIGIN, SYNC_URL, WSS_URL } from "../src/shared/config";

describe("R31 outbound allowlist", () => {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(__dirname, "..", "manifest.json"), "utf8"),
  ) as { host_permissions: string[]; declarative_net_request?: unknown };

  const rules = JSON.parse(
    fs.readFileSync(path.join(__dirname, "..", "src", "dnr", "rules.json"), "utf8"),
  ) as Array<{
    id: number;
    action: { type: string };
    condition: { requestDomains?: string[]; urlFilter?: string };
  }>;

  it("manifest host_permissions equals the shared allowlist (no silent extras)", () => {
    expect([...manifest.host_permissions].sort()).toEqual([...ALL_HOST_PERMISSIONS].sort());
  });

  it("build-time API/sync/ws URLs stay on allowlisted hostnames", () => {
    expect(isAllowedApiUrl(API_ORIGIN)).toBe(true);
    expect(isAllowedApiUrl(SYNC_URL)).toBe(true);
    expect(isAllowedApiUrl(WSS_URL.replace(/^ws/, "http"))).toBe(true);
  });

  it("rejects non-allowlisted API hosts", () => {
    expect(isAllowedApiUrl("https://evil.example/v1/ext/sync")).toBe(false);
    expect(isAllowedApiUrl("https://api.shopass.cyberskill.world.evil/x")).toBe(false);
    expect(isAllowedApiUrl("http://127.0.0.1:9999/v1")).toBe(false);
  });

  it("DNR rules only allow allowlisted API hosts (no marketplace block/redirect)", () => {
    expect(rules.length).toBeGreaterThan(0);
    expect(rules.length).toBeLessThanOrEqual(5);
    for (const rule of rules) {
      expect(rule.action.type).toBe("allow");
      const domains = rule.condition.requestDomains ?? [];
      for (const d of domains) {
        expect(ALLOWED_API_HOSTNAMES as readonly string[]).toContain(d);
      }
      if (rule.condition.urlFilter) {
        expect(rule.condition.urlFilter).toMatch(/127\.0\.0\.1:8080/);
        expect(rule.condition.urlFilter).not.toMatch(/shopee|tiktok|lazada/i);
      }
    }
  });

  it("declarative_net_request is declared when rules are non-empty", () => {
    expect(manifest.declarative_net_request).toBeDefined();
  });
});
