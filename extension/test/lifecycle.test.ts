import { getState, setState, defaultState } from "../src/shared/storage";
import { fakeChromeStorage } from "./helpers";

beforeEach(() => {
  globalThis.chrome = fakeChromeStorage() as unknown as typeof chrome;
});

test("state survive SW kill (rehydrate tu storage)", async () => {
  await setState({ ...defaultState(), lastSyncAt: 111 });

  jest.resetModules();
  const { getState: getState2 } = await import("../src/shared/storage");
  const s = await getState2();
  expect(s.lastSyncAt).toBe(111);
});

test("khong co state trong bien global", async () => {
  const fs = await import("fs");
  const src = fs.readFileSync("src/background/service-worker.ts", "utf8");
  expect(src).not.toMatch(/let\s+\w+State\s*=/);
});
