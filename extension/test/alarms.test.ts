import { registerAlarms, TICK } from "../src/background/alarms";

test("alarm chu ky >= 30s, khong setInterval", async () => {
  const created: { n: string; o: chrome.alarms.AlarmCreateInfo }[] = [];
  globalThis.chrome = {
    alarms: {
      create: (n: string, o: chrome.alarms.AlarmCreateInfo) =>
        created.push({ n, o }),
    },
  } as unknown as typeof chrome;

  registerAlarms();
  expect(created[0].n).toBe(TICK);
  expect(created[0].o.periodInMinutes).toBeGreaterThanOrEqual(0.5);

  const fs = await import("fs");
  const swSrc = fs.readFileSync("src/background/alarms.ts", "utf8");
  expect(swSrc).not.toMatch(/setInterval\(/);
});
