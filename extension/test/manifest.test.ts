import { promises as fs } from "fs";

test("manifest MV3, không <all_urls>, permissions tối thiểu", async () => {
  const m = JSON.parse(await fs.readFile("manifest.json", "utf8"));
  expect(m.manifest_version).toBe(3);
  expect(m.background.service_worker).toBeTruthy();
  expect(m.background.page).toBeUndefined();
  expect(JSON.stringify(m.host_permissions)).not.toMatch(/all_urls|:\/\/\*\/\*/);
  expect([...m.permissions].sort()).toEqual(["alarms", "storage"]);
});
