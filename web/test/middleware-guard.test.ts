/**
 * @jest-environment node
 */
import { middleware } from "../middleware";
import { NextRequest, NextResponse } from "next/server";

describe("Middleware Guard", () => {
  it("should redirect to /login if no refresh token on protected route", () => {
    const req = new NextRequest("http://localhost/dashboard");

    const res = middleware(req);

    expect(res).toBeDefined();
    expect(res?.status).toBe(307);
    expect(res?.headers.get("Location")).toBe("http://localhost/login?next=%2Fdashboard");
  });

  it("should allow request if refresh token exists on protected route", () => {
    const req = new NextRequest("http://localhost/dashboard");
    req.cookies.set("sandeal_refresh", "valid-token");

    const res = middleware(req);

    // NextResponse.next() returns a response with specific headers,
    // we just check that it didn't redirect (status 307)
    expect(res?.status).not.toBe(307);
  });

  it("should allow request to public routes without token", () => {
    const req = new NextRequest("http://localhost/login");

    const res = middleware(req);

    expect(res?.status).not.toBe(307);
  });

  it.each(["/wishlist", "/alerts", "/products/100/chart"])("protects %s", (path) => {
    const req = new NextRequest(`http://localhost${path}`);
    const res = middleware(req);

    expect(res?.status).toBe(307);
    expect(res?.headers.get("Location")).toBe(`http://localhost/login?next=${encodeURIComponent(path)}`);
  });

  it("keeps browser-capture parameters through login", () => {
    const path = "/capture?url=https%3A%2F%2Fshopee.vn%2Fitem-i.1.2&price=6490000";
    const req = new NextRequest(`http://localhost${path}`);

    const res = middleware(req);

    expect(res?.status).toBe(307);
    expect(res?.headers.get("Location")).toBe(`http://localhost/login?next=${encodeURIComponent(path)}`);
  });

  it("protects the bookmarklet installation guide", () => {
    const req = new NextRequest("http://localhost/capture-guide");
    const res = middleware(req);

    expect(res?.status).toBe(307);
    expect(res?.headers.get("Location")).toBe("http://localhost/login?next=%2Fcapture-guide");
  });
});
