import { NextResponse } from "next/server";
import { cookies } from "next/headers";

export async function POST() {
  const cookieStore = cookies();
  const refreshToken = cookieStore.get("refresh_token")?.value;

  if (!refreshToken) {
    return NextResponse.json({ error: "Missing refresh token" }, { status: 401 });
  }

  const BASE = process.env.NEXT_PUBLIC_API_BASE_URL || "";
  
  try {
    const res = await fetch(`${BASE}/v1/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!res.ok) {
      // Clear cookie on failure
      cookieStore.delete("refresh_token");
      return NextResponse.json({ error: "Failed to refresh" }, { status: 401 });
    }

    const data = await res.json();
    
    // Set new refresh token in httpOnly cookie
    if (data.refresh_token) {
      cookieStore.set("refresh_token", data.refresh_token, {
        httpOnly: true,
        secure: process.env.NODE_ENV === "production",
        sameSite: "lax",
        path: "/",
      });
    }

    return NextResponse.json({ accessToken: data.access_token });
  } catch (error) {
    return NextResponse.json({ error: "Internal error" }, { status: 500 });
  }
}
