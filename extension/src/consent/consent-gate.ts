/**
 * consent-gate.ts — Chặn đường dữ liệu trước mọi đọc/gửi (DEC-EXT-33).
 * Gate đọc consent state từ chrome.storage (bền, NFR-EXT-001).
 * Rút consent có hiệu lực ngay — gate re-read state mỗi lần gọi.
 */
import type { ConsentPurpose } from "../shared/types";
import { getConsent } from "./consent-store";

/**
 * ensureConsent — kiểm consent cho một mục cụ thể.
 * Trả false nếu chưa đồng ý → đường dữ liệu MUST bị chặn.
 */
export async function ensureConsent(purpose: ConsentPurpose): Promise<boolean> {
  const rec = await getConsent();
  return rec.granted.includes(purpose);
}
