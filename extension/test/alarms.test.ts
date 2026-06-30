import { registerAlarms, TICK } from "../src/background/alarms";
import { promises as fs } from "fs";

test("alarm chu kỳ >= 30s, không setInterval", async () => {
  const created: any[] = [];
  (globalThis as any).chrome = { alarms: { create: (n: string, o: any) => created.push({ n, o }) } } as any;
  registerAlarms();
  expect(created[0].n).toBe(TICK);
  expect(created[0].o.periodInMinutes).toBeGreaterThanOrEqual(0.5); // >=30s

  const swSrc = await fs.readFile("src/background/alarms.ts", "utf8");
  expect(swSrc).not.toMatch(/setInterval\(/);
});
