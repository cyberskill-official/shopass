import { CoinTask } from "../shared/types";

// Hiện checklist + nút DẪN tới trang nhiệm vụ trên sàn; user TỰ bấm (DEC-CART-32).
export function renderChecklist(tasks: CoinTask[], container: HTMLElement) {
  container.innerHTML = "";
  
  if (tasks.length === 0) {
    const p = document.createElement("p");
    p.textContent = "Chưa lấy được trạng thái xu.";
    container.appendChild(p);
    return;
  }

  for (const t of tasks) {
    const row = document.createElement("div");
    row.className = "checklist-row";
    row.textContent = `${t.taskType}: ${t.done ? "Đã xong" : "Chưa làm"}`;
    
    if (!t.done) {
      const a = document.createElement("a");
      a.href = taskUrlOnPlatform(t.taskType);
      a.textContent = " Tới trang nhiệm vụ";
      a.target = "_blank";
      row.appendChild(a);
      // KHÔNG có nút "tự động hoàn thành"
    }
    
    container.appendChild(row);
  }
  
  const note = document.createElement("div");
  note.className = "note";
  note.textContent = "Shopass chỉ nhắc; bạn tự thực hiện trên sàn.";
  container.appendChild(note);
}

function taskUrlOnPlatform(taskType: string): string {
  // Mock function to return a link to the task page based on type
  if (taskType === "daily_checkin") {
    return "/coin/checkin";
  }
  return "/coin";
}
