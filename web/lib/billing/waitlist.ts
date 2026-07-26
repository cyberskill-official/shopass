export type WaitlistPayload = {
  email: string;
  zalo?: string;
  source?: string;
  tier_interest?: "premium_basic" | "premium_plus" | "premium_pro";
};

export async function submitWaitlist(payload: WaitlistPayload): Promise<{ ok: boolean; id?: number }> {
  const res = await fetch("/api/waitlist", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const data = (await res.json().catch(() => ({}))) as { ok?: boolean; id?: number; error?: string };
  if (!res.ok) {
    throw new Error(data.error || "Không gửi được đăng ký");
  }
  return { ok: Boolean(data.ok), id: data.id };
}
