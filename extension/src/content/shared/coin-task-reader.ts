import { CoinTask } from "../../shared/types";

// readCoinTasks CHỈ đọc trạng thái nhiệm vụ xu hiển thị (DEC-CART-31).
// TUYỆT ĐỐI KHÔNG click, KHÔNG dispatch sự kiện, KHÔNG gọi API hoàn thành.
export function readCoinTasks(): CoinTask[] {
  const rows = document.querySelectorAll(".coin-task-row"); 
  const tasks: CoinTask[] = [];

  for (const el of rows) {
    tasks.push({
      taskType: parseTaskType(el),
      done: parseDoneState(el),
    });
  }

  return tasks;
}

function parseTaskType(el: Element): string {
  const title = el.querySelector(".task-title")?.textContent?.toLowerCase() || "";
  if (title.includes("điểm danh") || title.includes("checkin") || title.includes("check in")) {
    return "daily_checkin";
  }
  if (title.includes("live")) {
    return "watch_live";
  }
  if (title.includes("video")) {
    return "watch_video";
  }
  return "other";
}

function parseDoneState(el: Element): boolean {
  const btn = el.querySelector("button, .btn-task");
  if (!btn) return false;
  const text = btn.textContent?.toLowerCase() || "";
  // Nếu hiển thị là đã làm/hoàn thành/done thì return true
  if (text.includes("đã") || text.includes("hoàn thành") || text.includes("xong") || text.includes("claimed")) {
    return true;
  }
  return false;
}
