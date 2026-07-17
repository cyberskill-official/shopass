"use client";

import { useCallback, useEffect, useState } from "react";
import { getAccessToken, tryRefreshOnce, type RefreshResult } from "@/lib/auth";

type SessionState = "checking" | "ready" | "unavailable";

function loginRedirect(): void {
  const next = `${window.location.pathname}${window.location.search}${window.location.hash}`;
  const destination = new URL("/login", window.location.origin);
  destination.searchParams.set("next", next);
  window.location.replace(destination.toString());
}

// Middleware can only verify that a host-only refresh cookie exists. This
// client boundary validates/rotates it before rendering protected application
// content, so a revoked or expired cookie does not act like an active session.
export function SessionGuard({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<SessionState>("checking");

  const establishSession = useCallback(async () => {
    setState("checking");

    // A successful login just placed a short-lived access token in module
    // memory. Do not immediately rotate its refresh token a second time.
    if (getAccessToken()) {
      setState("ready");
      return;
    }

    const result: RefreshResult = await tryRefreshOnce();
    if (result === "refreshed") {
      setState("ready");
      return;
    }
    if (result === "invalid") {
      loginRedirect();
      return;
    }
    setState("unavailable");
  }, []);

  useEffect(() => {
    void establishSession();
  }, [establishSession]);

  if (state === "ready") return <>{children}</>;

  if (state === "unavailable") {
    return (
      <main className="mx-auto flex min-h-[50vh] max-w-lg flex-col items-center justify-center gap-4 px-6 text-center">
        <h1 className="text-xl font-bold text-slate-900">Không thể xác thực phiên đăng nhập</h1>
        <p className="text-sm leading-6 text-slate-600">
          Dịch vụ đăng nhập đang tạm thời không phản hồi. Phiên của bạn chưa bị đăng xuất.
        </p>
        <button
          type="button"
          onClick={() => void establishSession()}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-700"
        >
          Thử lại
        </button>
      </main>
    );
  }

  return (
    <main className="mx-auto flex min-h-[50vh] max-w-lg items-center justify-center px-6 text-center text-sm text-slate-600">
      Đang xác thực phiên đăng nhập…
    </main>
  );
}
