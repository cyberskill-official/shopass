import { tryRefreshOnce, logout, getAccessToken } from "./auth";

class UnauthorizedError extends Error {
  constructor(message = "Unauthorized") {
    super(message);
    this.name = "UnauthorizedError";
  }
}

class AuthenticationUnavailableError extends Error {
  constructor(message = "Authentication temporarily unavailable") {
    super(message);
    this.name = "AuthenticationUnavailableError";
  }
}

export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const BASE = process.env.NEXT_PUBLIC_API_BASE_URL || "";
  const headers = new Headers(init.headers);
  const accessToken = getAccessToken();
  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);
  let res = await fetch(`${BASE}${path}`, { ...init, headers });

  if (res.status === 401) {
    const refreshResult = await tryRefreshOnce();
    if (refreshResult === "invalid") {
      await logout();
      throw new UnauthorizedError();
    }
    if (refreshResult !== "refreshed") {
      throw new AuthenticationUnavailableError();
    }
    const newAccessToken = getAccessToken();
    if (newAccessToken) {
      headers.set("Authorization", `Bearer ${newAccessToken}`);
    } else {
      headers.delete("Authorization");
    }
    res = await fetch(`${BASE}${path}`, { ...init, headers });
  }
  return res;
}
