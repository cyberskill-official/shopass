"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setMessage("");
    setSubmitting(true);
    try {
      const res = await fetch("/api/auth/password/reset-request", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });
      const body = await res.json().catch(() => null);
      if (!res.ok) {
        setError(body?.error || "Không gửi được yêu cầu");
        return;
      }
      setMessage(
        typeof body?.message === "string"
          ? body.message
          : "Nếu tài khoản tồn tại, hướng dẫn đặt lại mật khẩu đã được gửi.",
      );
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
        <h1 className="mt-6 text-3xl font-black tracking-tight text-slate-950">Quên mật khẩu</h1>
        <p className="mt-3 text-sm leading-relaxed text-slate-600">
          Nhập email đã đăng ký. Phản hồi luôn giống nhau (không lộ tài khoản có tồn tại hay không).
        </p>
        <p className="mt-2 rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-950">
          Closed beta: email reset chỉ đến hộp thư khi SMTP đã cấu hình (R23). Nếu chưa có SMTP, yêu cầu
          vẫn được ghi nhận an toàn nhưng không có thư gửi đi.
        </p>

        {error && (
          <div role="alert" className="mt-4 rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-800">
            {error}
          </div>
        )}
        {message && (
          <div role="status" className="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-900">
            {message}
          </div>
        )}

        <form onSubmit={onSubmit} className="mt-6 space-y-4">
          <label className="block text-xs font-extrabold uppercase tracking-wide text-slate-500">
            Email
            <input
              required
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="mt-2 w-full rounded-xl border border-slate-200 px-4 py-3 text-sm font-medium"
              autoComplete="email"
            />
          </label>
          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-xl bg-slate-950 py-3.5 text-sm font-extrabold text-white disabled:opacity-60"
          >
            {submitting ? "Đang gửi…" : "Gửi hướng dẫn"}
          </button>
        </form>
      </div>
    </div>
  );
}
