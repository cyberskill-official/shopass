/** @jest-environment jsdom */
import { readFile } from "fs/promises";
import { join } from "path";
import { readCoinTasks } from "../src/content/shared/coin-task-reader";

test("reader + UI KHÔNG auto-click / dispatch / gọi API hoàn thành", async () => {
  for (const f of ["src/content/shared/coin-task-reader.ts", "src/ui/coin-checklist.ts"]) {
    const src = await readFile(join(__dirname, "..", f), "utf8");
    expect(src).not.toMatch(/\.click\s*\(/);        // không mô phỏng click
    expect(src).not.toMatch(/dispatchEvent/);       // không phát sự kiện chuột
    expect(src).not.toMatch(/complete.?task|claim.?coin|hoàn thành nhiệm vụ/i); // không gọi API hoàn thành
  }
});

test("reader chỉ đọc trạng thái, KHÔNG chạm credential", async () => {
  const src = await readFile(join(__dirname, "../src/content/shared/coin-task-reader.ts"), "utf8");
  expect(src).not.toMatch(/document\.cookie/);
  expect(src).not.toMatch(/Authorization|token/i);
});

test("readCoinTasks trả trạng thái done không đổi DOM", () => {
  document.body.innerHTML = `
    <div class="coin-task-row">
      <div class="task-title">Điểm danh</div>
      <button class="btn-task">Đã xong</button>
    </div>
    <div class="coin-task-row">
      <div class="task-title">Xem live</div>
      <button class="btn-task">Làm ngay</button>
    </div>
  `;
  const before = document.body.innerHTML;
  const tasks = readCoinTasks();
  expect(tasks).toHaveLength(2);
  expect(tasks.some((t) => t.done)).toBe(true);
  expect(tasks[0].taskType).toBe("daily_checkin");
  expect(tasks[0].done).toBe(true);
  expect(tasks[1].taskType).toBe("watch_live");
  expect(tasks[1].done).toBe(false);
  expect(document.body.innerHTML).toBe(before); // đọc không sửa DOM (không click)
});
