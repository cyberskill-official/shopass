import { apiFetch } from "@/lib/api";

export type ReferralMe = {
  code: string;
  uses: number;
  has_referrer: boolean;
  reward_note: string;
};

export async function getReferralMe(): Promise<ReferralMe> {
  const res = await apiFetch("/v1/referral/me");
  if (!res.ok) throw new Error("Không tải được mã giới thiệu");
  return (await res.json()) as ReferralMe;
}

export async function attributeReferral(code: string): Promise<void> {
  const res = await apiFetch("/v1/referral/attribute", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ code }),
  });
  if (res.ok) return;
  const body = (await res.json().catch(() => ({}))) as { error?: string };
  if (body.error === "self_referral") throw new Error("Không thể tự giới thiệu");
  if (body.error === "already_attributed") throw new Error("Bạn đã được gắn mã giới thiệu");
  if (body.error === "unknown_code") throw new Error("Mã giới thiệu không hợp lệ");
  throw new Error("Không gắn được mã giới thiệu");
}

const PENDING_KEY = "shopass_pending_ref";

export function capturePendingReferral(code: string | null | undefined): void {
  if (!code || typeof window === "undefined") return;
  const cleaned = code.trim().toUpperCase();
  if (cleaned.length < 4 || cleaned.length > 32) return;
  window.localStorage.setItem(PENDING_KEY, cleaned);
}

export function readPendingReferral(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(PENDING_KEY);
}

export function clearPendingReferral(): void {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(PENDING_KEY);
}
