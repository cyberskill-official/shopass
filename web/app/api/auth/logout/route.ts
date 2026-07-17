import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { gatewayFetch, readRefreshCookie, refreshCookieName, requestHasAllowedOrigin } from "@/lib/server-auth";

export async function POST(request: NextRequest) {
  if (!requestHasAllowedOrigin(request)) {
    return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
  }
  const refreshToken = readRefreshCookie(request);
  if (refreshToken) {
    await gatewayFetch("/v1/auth/logout", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    }, request).catch(() => undefined);
  }
  const cookieStore = await cookies();
  cookieStore.delete(refreshCookieName);
  cookieStore.delete("sandeal_refresh");
  cookieStore.delete("refresh_token");

  return NextResponse.json({ success: true });
}
