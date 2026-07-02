/**
 * @jest-environment node
 */
import { middleware } from "../middleware";
import { NextRequest, NextResponse } from "next/server";

describe("Middleware Guard", () => {
  it("should redirect to /login if no refresh_token on protected route", () => {
    const req = new NextRequest("http://localhost/dashboard");
    
    const res = middleware(req);
    
    expect(res).toBeDefined();
    expect(res?.status).toBe(307);
    expect(res?.headers.get("Location")).toBe("http://localhost/login?next=%2Fdashboard");
  });

  it("should allow request if refresh_token exists on protected route", () => {
    const req = new NextRequest("http://localhost/dashboard");
    req.cookies.set("refresh_token", "valid-token");
    
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
});
