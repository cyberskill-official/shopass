import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { gatewayFetch, readRefreshCookie, refreshCookieName, refreshCookieOptions, requestHasAllowedOrigin } from "@/lib/server-auth";

async function clearRefreshCookies() {
  const cookieStore = await cookies();
  cookieStore.delete(refreshCookieName);
  // Clean up names used by pre-production/local builds, but never trust either
  // one in production (see readRefreshCookie).
  cookieStore.delete("sandeal_refresh");
  cookieStore.delete("refresh_token");
}

export async function POST(request: NextRequest) {
  if (!requestHasAllowedOrigin(request)) {
    return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
  }
  const refreshToken = readRefreshCookie(request);

  if (!refreshToken) {
    return NextResponse.json({ error: "Missing refresh token" }, { status: 401 });
  }

  try {
    const res = await gatewayFetch("/v1/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    }, request);

    if (!res.ok) {
      // An explicit invalid-token response is terminal. Keep a valid cookie on
      // transient 429/5xx failures so a Redis/JWKS/database outage does not
      // force every user to sign in again.
      if (res.status === 401) {
        await clearRefreshCookies();
        return NextResponse.json({ error: "Failed to refresh" }, { status: 401 });
      }
      return NextResponse.json(
        { error: "Authentication service temporarily unavailable" },
        { status: res.status },
      );
    }

    const data = await res.json();
    if (typeof data?.access_token !== "string" || typeof data?.refresh_token !== "string") {
      return NextResponse.json({ error: "Invalid auth response" }, { status: 502 });
    }
    (await cookies()).set(refreshCookieName, data.refresh_token, refreshCookieOptions);

    return NextResponse.json({ accessToken: data.access_token });
  } catch {
    return NextResponse.json(
      { error: "Authentication service temporarily unavailable" },
      { status: 503 },
    );
  }
}
