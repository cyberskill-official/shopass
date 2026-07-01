/**
 * redact.ts — quét giá trị chuỗi tìm dấu hiệu credential/PII.
 * Bịt kẽ hở "nhồi credential vào trường được phép" (§1 #6).
 */

// Patterns nhận diện credential/PII nhồi vào giá trị trường
const COOKIE_LIKE = /^(SPC_|SHOPEE_|session[_-]?id|token|jwt|Bearer)/i;
const LONG_TOKEN = /^[A-Za-z0-9+/=_-]{40,}$/; // chuỗi base64-like ≥40 ký tự
const EMAIL_LIKE = /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/;
const PHONE_VN = /^(?:\+?84|0)\d{9,10}$/;

/**
 * looksLikeCredential — true nếu giá trị trông giống credential hoặc PII.
 */
export function looksLikeCredential(val: unknown): boolean {
  if (typeof val !== "string") return false;
  if (COOKIE_LIKE.test(val)) return true;
  if (LONG_TOKEN.test(val)) return true;
  if (EMAIL_LIKE.test(val)) return true;
  if (PHONE_VN.test(val)) return true;
  return false;
}

/**
 * containsPiiOrCredential — quét toàn bộ payload tìm giá trị nghi PII/credential.
 * Trả true nếu phát hiện (phải từ chối cả payload).
 */
export function containsPiiOrCredential(payload: Record<string, unknown>): boolean {
  return scanObj(payload);
}

function scanObj(obj: unknown): boolean {
  if (typeof obj === "string") return looksLikeCredential(obj);
  if (Array.isArray(obj)) return obj.some(scanObj);
  if (obj && typeof obj === "object") {
    return Object.values(obj as Record<string, unknown>).some(scanObj);
  }
  return false;
}
