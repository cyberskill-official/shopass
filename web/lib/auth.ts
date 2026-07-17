export type RefreshResult = "refreshed" | "invalid" | "transient";

let refreshInFlight: Promise<RefreshResult> | null = null;

async function refreshAccessToken(): Promise<RefreshResult> {
  try {
    const res = await fetch("/api/auth/refresh", { method: "POST" });
    if (!res.ok) return res.status === 401 ? "invalid" : "transient";
    const data = await res.json();
    if (typeof data?.accessToken === "string" && data.accessToken) {
      setAccessToken(data.accessToken);
      return "refreshed";
    }
  } catch {
    // Preserve the refresh cookie after a network/JSON failure. apiFetch can
    // surface a recoverable outage without turning it into a logout.
  }
  return "transient";
}

// Multiple client requests commonly receive 401 together after a page reload.
// Coalesce their refresh attempt so a one-time refresh token is never rotated
// concurrently by the same browser session.
export function tryRefreshOnce(): Promise<RefreshResult> {
  if (!refreshInFlight) {
    refreshInFlight = refreshAccessToken().finally(() => {
      refreshInFlight = null;
    });
  }
  return refreshInFlight;
}

export async function logout(): Promise<void> {
  setAccessToken(null);
  await fetch("/api/auth/logout", { method: "POST" });
}

let accessToken: string | null = null;
export function setAccessToken(t: string | null) { accessToken = t; }
export function getAccessToken(): string | null { return accessToken; }
