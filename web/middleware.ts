import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { refreshCookieCandidates } from "@/lib/server-auth";

function readRefreshCookie(request: NextRequest): string | undefined {
  for (const name of refreshCookieCandidates()) {
    const value = request.cookies.get(name)?.value;
    if (value) return value;
  }
  return undefined;
}

export function middleware(request: NextRequest) {
  const protectedPrefixes = ["/dashboard", "/wishlist", "/alerts", "/products"];
  if (protectedPrefixes.some((prefix) => request.nextUrl.pathname.startsWith(prefix))) {
    const refreshToken = readRefreshCookie(request);

    if (!refreshToken) {
      const loginUrl = new URL("/login", request.url);
      loginUrl.searchParams.set("next", request.nextUrl.pathname);
      // HTTP 307 Temporary Redirect as required by DEC-WEB-05
      return NextResponse.redirect(loginUrl, 307);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
};
