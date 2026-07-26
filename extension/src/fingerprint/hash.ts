/** One-way client device hash — never send raw UA/canvas/fonts (TRUST-006). */

function toHex(buf: ArrayBuffer): string {
  return [...new Uint8Array(buf)].map((b) => b.toString(16).padStart(2, "0")).join("");
}

export async function hashDeviceSignals(signals: string[]): Promise<string> {
  const material = signals.filter(Boolean).join("|");
  const data = new TextEncoder().encode(material);
  const digest = await crypto.subtle.digest("SHA-256", data);
  return toHex(digest);
}
