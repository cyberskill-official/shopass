/** Next marketplace double-date (1.1 … 12.12) on or after `from` (UTC day). */
export function nextDoubleDate(from: Date = new Date()): { date: Date; label: string } {
  const start = new Date(Date.UTC(from.getUTCFullYear(), from.getUTCMonth(), from.getUTCDate()));
  for (let y = start.getUTCFullYear(); y <= start.getUTCFullYear() + 1; y++) {
    for (let m = 1; m <= 12; m++) {
      const d = new Date(Date.UTC(y, m - 1, m));
      if (d.getTime() >= start.getTime()) {
        return { date: d, label: `${m}.${m}` };
      }
    }
  }
  const fallback = new Date(Date.UTC(start.getUTCFullYear() + 1, 0, 1));
  return { date: fallback, label: "1.1" };
}

export function daysUntil(target: Date, from: Date = new Date()): number {
  const a = Date.UTC(from.getUTCFullYear(), from.getUTCMonth(), from.getUTCDate());
  const b = Date.UTC(target.getUTCFullYear(), target.getUTCMonth(), target.getUTCDate());
  return Math.max(0, Math.round((b - a) / 86_400_000));
}
