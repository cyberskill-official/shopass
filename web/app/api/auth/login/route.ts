import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { gatewayFetch, refreshCookieName, refreshCookieOptions, requestHasAllowedOrigin } from "@/lib/server-auth";

export async function POST(request: NextRequest) {
	if (!requestHasAllowedOrigin(request)) {
		return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
	}

  try {
    const body = await request.json();
    if (typeof body?.email !== "string" || typeof body?.password !== "string") {
      return NextResponse.json({ error: "Invalid credentials" }, { status: 400 });
    }

    const upstream = await gatewayFetch("/v1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: body.email, password: body.password }),
    }, request);
    if (!upstream.ok) {
      return NextResponse.json({ error: "Invalid email or password" }, { status: upstream.status });
    }
    const data = await upstream.json();
    if (typeof data?.refresh_token !== "string" || typeof data?.access_token !== "string") {
      return NextResponse.json({ error: "Invalid auth response" }, { status: 502 });
    }

    const cookieStore = await cookies();
    cookieStore.set(refreshCookieName, data.refresh_token, refreshCookieOptions);

    // Refresh tokens never leave this server route. The access token is short
    // lived and held only in the browser's module memory.
    return NextResponse.json({ accessToken: data.access_token, expiresIn: data.expires_in });
  } catch {
    return NextResponse.json({ error: "Internal error" }, { status: 500 });
  }
}
