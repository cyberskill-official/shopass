import { NextRequest, NextResponse } from "next/server";
import { gatewayFetch, requestHasAllowedOrigin } from "@/lib/server-auth";

export async function POST(request: NextRequest) {
  if (!requestHasAllowedOrigin(request)) {
    return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
  }
  try {
    const body = await request.json();
    if (typeof body?.email !== "string" || typeof body?.password !== "string") {
      return NextResponse.json({ error: "Invalid registration" }, { status: 400 });
    }
    const upstream = await gatewayFetch("/v1/auth/register", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: body.email, password: body.password }),
    }, request);
    if (!upstream.ok) {
      const error = await upstream.json().catch(() => null);
      return NextResponse.json({ error: error?.error || "Không thể tạo tài khoản" }, { status: upstream.status });
    }
    return NextResponse.json({ success: true }, { status: 201 });
  } catch {
    return NextResponse.json({ error: "Internal error" }, { status: 500 });
  }
}
