import * as fs from "fs";

test("manifest MV3, khong all_urls, permissions toi thieu", () => {
  const m = JSON.parse(fs.readFileSync("manifest.json", "utf8"));
  expect(m.manifest_version).toBe(3);
  expect(m.background.service_worker).toBeTruthy();
  expect(m.background.page).toBeUndefined();
  expect(JSON.stringify(m.host_permissions)).not.toMatch(
    /all_urls|:\/\/\*\/\*/
  );
  expect(m.permissions.sort()).toEqual(["alarms", "storage"]);
});
