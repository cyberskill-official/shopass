import { NextRequest, NextResponse } from "next/server";
import { gatewayFetch, requestHasAllowedOrigin } from "@/lib/server-auth";

export async function POST(request: NextRequest) {
  if (!requestHasAllowedOrigin(request)) {
    return NextResponse.json({ error: "Invalid origin" }, { status: 403 });
  }

  try {
    const body = await request.json();
    const identifier =
      typeof body?.identifier === "string"
        ? body.identifier
        : typeof body?.email === "string"
          ? body.email
          : null;
    if (!identifier) {
      return NextResponse.json({ error: "Email required" }, { status: 400 });
    }

    const upstream = await gatewayFetch(
      "/v1/auth/password/reset-request",
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ identifier }),
      },
      request,
    );
    const data = await upstream.json().catch(() => null);
    if (!upstream.ok) {
      return NextResponse.json(
        { error: data?.error || "Không thể yêu cầu đặt lại mật khẩu" },
        { status: upstream.status },
      );
    }
    return NextResponse.json({
      message:
        typeof data?.message === "string"
          ? data.message
          : "Nếu tài khoản tồn tại, hướng dẫn đặt lại mật khẩu đã được gửi.",
    });
  } catch {
    return NextResponse.json({ error: "Internal error" }, { status: 500 });
  }
}
