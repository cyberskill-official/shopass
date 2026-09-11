import { NextRequest, NextResponse } from "next/server";
import { gatewayFetch, requestHasAllowedOrigin } from "@/lib/server-auth";

export async function POST(request: NextRequest) {
  if (!requestHasAllowedOrigin(request)) {
    return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
  }

  try {
    const body = await request.json();
    if (typeof body?.token !== "string" || typeof body?.new_password !== "string") {
      return NextResponse.json({ error: "Token và mật khẩu mới bắt buộc" }, { status: 400 });
    }

    const upstream = await gatewayFetch(
      "/v1/auth/password/reset-confirm",
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          token: body.token,
          new_password: body.new_password,
        }),
      },
      request,
    );
    const data = await upstream.json().catch(() => null);
    if (!upstream.ok) {
      return NextResponse.json(
        { error: data?.error || "Không thể đặt lại mật khẩu" },
        { status: upstream.status },
      );
    }
    return NextResponse.json({ status: "password_updated" });
  } catch {
    return NextResponse.json({ error: "Internal error" }, { status: 500 });
  }
}
