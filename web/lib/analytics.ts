/**
 * Funnel events stub until R40 (GA4 vs Plausible decision).
 * Safe no-op in production; tests can spy on trackEvent.
 */
export type AnalyticsEvent =
  | "install-click"
  | "signup-click"
  | "email-submit"
  | "how-click"
  | "trust-click";

export type AnalyticsPayload = Record<string, string | number | boolean | undefined>;

const buffer: { name: AnalyticsEvent; payload?: AnalyticsPayload }[] = [];

export function trackEvent(name: AnalyticsEvent, payload?: AnalyticsPayload): void {
  buffer.push({ name, payload });
  if (typeof window !== "undefined") {
    window.dispatchEvent(new CustomEvent("shopass:analytics", { detail: { name, payload } }));
  }
}

/** Test helper — not for product UI. */
export function __analyticsBuffer(): typeof buffer {
  return buffer;
}

export function __resetAnalyticsBuffer(): void {
  buffer.length = 0;
}
