import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  // Guard the (app) route group, specifically we can check path prefixes or all paths excluding some
  // We'll protect anything that is not public. Let's protect /dashboard for now as per minimal setup.
  if (request.nextUrl.pathname.startsWith("/dashboard")) {
    const refreshToken = request.cookies.get("refresh_token")?.value;
    
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
