import { promises as fs } from "fs";

test("manifest MV3, không <all_urls>, permissions tối thiểu", async () => {
  const m = JSON.parse(await fs.readFile("manifest.json", "utf8"));
  expect(m.manifest_version).toBe(3);
  expect(m.background.service_worker).toBeTruthy();
  expect(m.background.page).toBeUndefined();
  expect(JSON.stringify(m.host_permissions)).not.toMatch(/all_urls|:\/\/\*\/\*/);
  expect([...m.permissions].sort()).toEqual([
    "alarms",
    "declarativeNetRequest",
    "offscreen",
    "storage",
  ]);
  // R31: audited non-empty DNR allowlist for Shopass API hosts.
  expect(m.declarative_net_request?.rule_resources?.length).toBeGreaterThan(0);
  expect(m.permissions).toContain("declarativeNetRequest");
});
