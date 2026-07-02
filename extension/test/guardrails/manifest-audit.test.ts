import { auditManifest } from "../../src/guardrails/manifest-audit";

test("cookies permission bị bắt", () => {
  expect(auditManifest({ permissions: ["cookies", "storage"] }).length).toBe(1);
});

test("manifest tối thiểu hợp lệ -> sạch", () => {
  expect(auditManifest({ permissions: ["storage", "alarms"],
    host_permissions: ["https://shopee.vn/*"] })).toHaveLength(0);
});

test("webRequestBlocking permission bị bắt", () => {
  expect(auditManifest({ permissions: ["webRequestBlocking"] }).length).toBe(1);
});
