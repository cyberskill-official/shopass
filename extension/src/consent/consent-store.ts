/**
 * consent-store.ts — Lưu ConsentRecord bền trong chrome.storage.
 * Mặc định granted: [] — KHÔNG mục nào bật (DEC-EXT-29: im lặng != đồng thuận).
 * Ghi record tái lập được (DEC-EXT-31): policyVersion + decidedAt + granted.
 */
import type { ConsentPurpose, ConsentRecord } from "../shared/types";

/** Phiên bản chính sách hiện tại — đổi khi disclosure cập nhật (§1 #9). */
export const POLICY_VERSION = "2026-06-27";

const KEY = "sandeal:consent";

/**
 * getConsent — đọc consent record từ chrome.storage.
 * Trả mặc định granted: [] nếu chưa ghi (cài mới).
 */
export async function getConsent(): Promise<ConsentRecord> {
  const o = await chrome.storage.local.get(KEY);
  const rec = o[KEY] as ConsentRecord | undefined;
  if (!rec) {
    return { policyVersion: POLICY_VERSION, decidedAt: 0, granted: [] };
  }
  // §1 #9: nếu policyVersion khác → consent cũ không tự áp cho version mới
  if (rec.policyVersion !== POLICY_VERSION) {
    return { policyVersion: POLICY_VERSION, decidedAt: 0, granted: [] };
  }
  return rec;
}

/**
 * setConsent — ghi consent record mới (bật/tắt mục).
 * Gửi TASK-COMPLY-001 để tái lập được.
 */
export async function setConsent(granted: ConsentPurpose[]): Promise<void> {
  const rec: ConsentRecord = {
    policyVersion: POLICY_VERSION,
    decidedAt: Date.now(),
    granted,
  };
  await chrome.storage.local.set({ [KEY]: rec });
  // TASK-COMPLY-001: report consent to central compliance framework
  // (implemented when TASK-COMPLY-001 integration is wired)
}
