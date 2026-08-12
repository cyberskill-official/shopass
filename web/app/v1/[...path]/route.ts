import { NextRequest, NextResponse } from "next/server";

export const dynamic = "force-dynamic";

type RouteCtx = { params: Promise<{ path: string[] }> };

const FORWARDED_REQUEST_HEADERS = [
  "authorization",
  "content-type",
  "accept",
  "x-request-id",
  "x-real-ip",
] as const;

const FORWARDED_RESPONSE_HEADERS = [
  "content-type",
  "x-request-id",
  "retry-after",
  "www-authenticate",
] as const;

function gatewayBase(): string {
  const raw = process.env.GATEWAY_INTERNAL_BASE_URL || "http://127.0.0.1:8080";
  return raw.replace(/\/+$/, "");
}

async function proxyToGateway(request: NextRequest, ctx: RouteCtx): Promise<NextResponse> {
  const { path } = await ctx.params;
  if (!path?.length || path[0] === "auth") {
    return new NextResponse("not found", { status: 404 });
  }

  const upstreamURL = `${gatewayBase()}/v1/${path.join("/")}${request.nextUrl.search}`;
  const headers = new Headers();
  for (const name of FORWARDED_REQUEST_HEADERS) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }

  const init: RequestInit = {
    method: request.method,
    headers,
    redirect: "manual",
  };
  if (request.method !== "GET" && request.method !== "HEAD") {
    init.body = await request.arrayBuffer();
  }

  let upstream: Response;
  try {
    upstream = await fetch(upstreamURL, init);
  } catch {
    return NextResponse.json({ error: "upstream unavailable" }, { status: 502 });
  }

  const responseHeaders = new Headers();
  for (const name of FORWARDED_RESPONSE_HEADERS) {
    const value = upstream.headers.get(name);
    if (value) responseHeaders.set(name, value);
  }
  return new NextResponse(upstream.body, {
    status: upstream.status,
    headers: responseHeaders,
  });
}

export async function GET(request: NextRequest, ctx: RouteCtx) {
  return proxyToGateway(request, ctx);
}

export async function POST(request: NextRequest, ctx: RouteCtx) {
  return proxyToGateway(request, ctx);
}

export async function PUT(request: NextRequest, ctx: RouteCtx) {
  return proxyToGateway(request, ctx);
}

export async function PATCH(request: NextRequest, ctx: RouteCtx) {
  return proxyToGateway(request, ctx);
}

export async function DELETE(request: NextRequest, ctx: RouteCtx) {
  return proxyToGateway(request, ctx);
}
