export async function tryRefreshOnce(): Promise<boolean> {
  const res = await fetch("/api/auth/refresh", { method: "POST" });
  if (!res.ok) return false;
  const data = await res.json();
  if (data.accessToken) {
    setAccessToken(data.accessToken);
    return true;
  }
  return false;
}

export async function logout(): Promise<void> {
  setAccessToken(null);
  await fetch("/api/auth/logout", { method: "POST" });
}

let accessToken: string | null = null;
export function setAccessToken(t: string | null) { accessToken = t; }
export function getAccessToken(): string | null { return accessToken; }
