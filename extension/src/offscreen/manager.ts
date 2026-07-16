/**
 * manager.ts — quản lý vòng đời offscreen: kiểm -> tạo -> dùng -> đóng NGAY.
 * MUST NOT để offscreen mở thường trực (DEC-EXT-19).
 */
import type { ParseDomRequest, ParseDomResult } from "../shared/types";

const OFFSCREEN_PATH = "offscreen/offscreen.html";
const TASK_TIMEOUT_MS = 10_000; // §1 #12: trần 10s, tự đóng nếu quá

/**
 * sendWithTimeout gửi message đến offscreen và trả kết quả.
 * Nếu quá timeout, reject để caller đóng offscreen.
 */
function sendWithTimeout(
  req: ParseDomRequest,
  timeoutMs: number
): Promise<ParseDomResult> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error("offscreen task timeout"));
    }, timeoutMs);

    chrome.runtime.sendMessage(req, (response: ParseDomResult) => {
      clearTimeout(timer);
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message));
      } else {
        resolve(response);
      }
    });
  });
}

/**
 * parseDomOffscreen — tạo offscreen nếu chưa có, gửi HTML thô, nhận kết quả,
 * rồi đóng offscreen NGAY (DEC-EXT-19).
 * Kết quả cần qua minimize (TASK-EXT-003) trước khi rời client.
 */
export async function parseDomOffscreen(
  req: ParseDomRequest
): Promise<ParseDomResult> {
  // §1 #2: kiểm hasDocument trước khi tạo
  const hasDoc = await chrome.offscreen.hasDocument();
  if (!hasDoc) {
    await chrome.offscreen.createDocument({
      url: OFFSCREEN_PATH,
      reasons: [chrome.offscreen.Reason.DOM_SCRAPING],
      justification: "Parse HTML giỏ hàng đã render ngoài service worker",
    });
  }

  try {
    const result = await sendWithTimeout(req, TASK_TIMEOUT_MS);
    return result;
  } finally {
    // §1 #3: đóng NGAY sau khi xong, không để mở thường trực
    try {
      await chrome.offscreen.closeDocument();
    } catch {
      // Already closed or never opened — ignore
    }
  }
}

export { OFFSCREEN_PATH, TASK_TIMEOUT_MS };
