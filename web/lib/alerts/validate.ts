export type RuleType = "price_below" | "drop_pct" | "real_sale" | "bottom_predicted";
export type Channel = "push" | "email" | "sms";
export const CHANNELS: Channel[] = ["push", "email", "sms"];

export function needsThreshold(rt: RuleType): boolean {
  return rt === "price_below" || rt === "drop_pct";
}

export function validateAlert(
  rt: RuleType,
  threshold: number | null,
  channels: Channel[]
): string | null {
  if (channels.length === 0) return "Chọn ít nhất một kênh";
  if (channels.some((c) => !CHANNELS.includes(c))) return "Kênh không hợp lệ";
  
  if (rt === "price_below") {
    if (threshold == null || !Number.isInteger(threshold) || threshold <= 0) {
      return "Giá phải là số nguyên dương (VND)";
    }
  } else if (rt === "drop_pct") {
    if (threshold == null || !Number.isInteger(threshold) || threshold < 1 || threshold > 99) {
      return "Phần trăm phải trong 1..99";
    }
  } else {
    if (threshold != null) {
      return "Loại này không nhận ngưỡng";
    }
  }
  
  return null;
}
