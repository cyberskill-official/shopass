import { getState, setState, defaultState } from "../src/shared/storage";
import { fakeChromeStorage } from "./helpers";
import { promises as fs } from "fs";

test("state survive SW kill (rehydrate từ storage)", async () => {
  (globalThis as any).chrome = { storage: fakeChromeStorage() };
  await setState({ ...defaultState(), lastSyncAt: 111 });

  // mô phỏng SW kill: vứt mọi biến module, đọc lại sạch từ storage
  jest.resetModules();
  const { getState: getState2 } = await import("../src/shared/storage");
  const s = await getState2();
  expect(s.lastSyncAt).toBe(111); // KHÔNG mất qua "kill"
});

test("không có state trong biến global", async () => {
  const src = await fs.readFile("src/background/service-worker.ts", "utf8");
  // entrypoint chỉ đăng ký listener + tạo alarm, không khai biến state
  expect(src).not.toMatch(/let\s+\w+State\s*=/);
});
