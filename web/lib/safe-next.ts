const defaultNextPath = "/dashboard";
/** Post-signup aha path (R45). Prefer this when `signup=1` and no explicit next. */
export const onboardingNextPath = "/onboarding";

// Login's `next` parameter is attacker-controlled. Normalize it with the URL
// parser and only retain same-origin locations, rather than relying on a
// prefix check (which can be bypassed with protocol-relative/backslash URLs).
export function safeNextPath(candidate: string | null | undefined, origin: string): string {
  if (!candidate) return defaultNextPath;

  try {
    const base = new URL(origin);
    const target = new URL(candidate, base);
    if (target.origin !== base.origin) return defaultNextPath;
    return `${target.pathname}${target.search}${target.hash}`;
  } catch {
    return defaultNextPath;
  }
}
