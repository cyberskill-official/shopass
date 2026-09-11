"use client";

import Link from "next/link";
import { FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";

export default function ResetPasswordPage() {
  const router = useRouter();
  const [token, setToken] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [done, setDone] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    setToken(params.get("token") ?? "");
  }, []);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      const res = await fetch("/api/auth/password/reset-confirm", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token, new_password: password }),
      });
      const body = await res.json().catch(() => null);
      if (!res.ok) {
        setError(body?.error || "Token không hợp lệ hoặc đã hết hạn");
        return;
      }
      setDone(true);
      setTimeout(() => router.push("/login"), 1500);
    } catch {
      setError("Không thể kết nối");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="mesh-bg flex min-h-screen items-center justify-center p-4">
      <div className="w-full max-w-md rounded-[2rem] border border-white/80 bg-white/90 p-8 shadow-xl backdrop-blur">
        <Link href="/login" className="text-sm font-bold text-slate-500 hover:text-slate-900">
          ← Đăng nhập
        </Link>
        <h1 className="mt-6 text-3xl font-black tracking-tight text-slate-950">Đặt mật khẩu mới</h1>
        <p className="mt-3 text-sm text-slate-600">
          Sau khi đổi mật khẩu, mọi phiên đăng nhập cũ bị thu hồi.
        </p>

        {error && (
          <div role="alert" className="mt-4 rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-800">
            {error}
          </div>
        )}
        {done && (
          <div role="status" className="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-900">
            Đã cập nhật. Chuyển tới đăng nhập…
          </div>
        )}

        <form onSubmit={onSubmit} className="mt-6 space-y-4">
          <label className="block text-xs font-extrabold uppercase tracking-wide text-slate-500">
            Token
            <input
              required
              value={token}
              onChange={(e) => setToken(e.target.value)}
              className="mt-2 w-full rounded-xl border border-slate-200 px-4 py-3 font-mono text-sm"
              autoComplete="off"
            />
          </label>
          <label className="block text-xs font-extrabold uppercase tracking-wide text-slate-500">
            Mật khẩu mới (tối thiểu 8 ký tự)
            <input
              required
              type="password"
              minLength={8}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="mt-2 w-full rounded-xl border border-slate-200 px-4 py-3 text-sm"
              autoComplete="new-password"
            />
          </label>
          <button
            type="submit"
            disabled={submitting || done}
            className="w-full rounded-xl bg-slate-950 py-3.5 text-sm font-extrabold text-white disabled:opacity-60"
          >
            {submitting ? "Đang lưu…" : "Lưu mật khẩu"}
          </button>
        </form>
      </div>
    </div>
  );
}
