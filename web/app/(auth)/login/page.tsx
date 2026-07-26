"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { setAccessToken } from "@/lib/auth";
import { onboardingNextPath, safeNextPath } from "@/lib/safe-next";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [nextPath, setNextPath] = useState(onboardingNextPath);
  const router = useRouter();
  const [isSignup, setIsSignup] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const signup = params.get("signup") === "1";
    setIsSignup(signup);
    const explicit = params.get("next");
    setNextPath(
      safeNextPath(explicit ?? (signup ? onboardingNextPath : null), window.location.origin),
    );
  }, []);

  const finishWithToken = (accessToken: string) => {
    setAccessToken(accessToken);
    router.push(nextPath);
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setMessage("");
    setSubmitting(true);

    try {
      if (isSignup) {
        const reg = await fetch("/api/auth/register", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
        });
        if (!reg.ok) {
          const body = await reg.json().catch(() => null);
          setError(body?.error || "Không thể tạo tài khoản");
          return;
        }
        // Auto-login so R45 time-to-first-alert stays under ~2 minutes.
        const login = await fetch("/api/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
        });
        if (!login.ok) {
          setIsSignup(false);
          setMessage("Tài khoản đã tạo. Đăng nhập để bắt đầu theo dõi giá.");
          return;
        }
        const data = await login.json();
        if (typeof data?.accessToken !== "string") {
          setError("Phản hồi đăng nhập không hợp lệ");
          return;
        }
        finishWithToken(data.accessToken);
        return;
      }

      const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => null);
        setError(body?.error || "Email hoặc mật khẩu không đúng");
        return;
      }
      const data = await res.json();
      if (typeof data?.accessToken !== "string") {
        setError("Phản hồi đăng nhập không hợp lệ");
        return;
      }
      finishWithToken(data.accessToken);
    } catch {
      setError("Không thể kết nối tới dịch vụ đăng nhập");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="mesh-bg flex min-h-screen items-center justify-center p-4 sm:p-6 lg:p-8">
      <div className="grid w-full max-w-5xl min-w-0 overflow-hidden rounded-[2rem] border border-white/80 bg-white/80 shadow-2xl shadow-indigo-100/50 backdrop-blur-xl lg:grid-cols-[1fr_1.1fr]">

        {/* Left Side: Brand Panel */}
        <div className="relative hidden flex-col justify-between overflow-hidden bg-slate-950 p-10 text-white lg:flex xl:p-12">
          <div className="absolute -left-10 -top-10 h-64 w-64 rounded-full bg-blue-600/20 blur-3xl" />
          <div className="absolute -bottom-10 -right-10 h-64 w-64 rounded-full bg-violet-600/20 blur-3xl" />

          <div className="relative z-10">
            <Link href="/" className="flex items-center gap-3 text-xl font-black transition-opacity hover:opacity-80">
              <span className="grid h-10 w-10 place-items-center rounded-2xl bg-gradient-to-br from-blue-500 to-violet-500 shadow-lg shadow-blue-500/20">S</span>
              <span>Shop<span className="text-blue-400">ass</span></span>
            </Link>

            <h1 className="mt-20 text-4xl font-black leading-[1.15] tracking-tight text-white xl:text-5xl">
              Đừng mua theo cảm tính.<br />
              <span className="text-blue-400">Mua theo dữ liệu.</span>
            </h1>

            <p className="mt-6 max-w-sm text-base leading-relaxed text-slate-400">
              Theo dõi lịch sử giá minh bạch để biết một deal thật sự tốt hay chỉ đang được nền tảng đánh lừa.
            </p>
          </div>

          <div className="relative z-10">
            <p className="inline-flex items-center gap-2 rounded-full bg-white/10 px-3 py-1.5 text-xs font-bold tracking-wide text-blue-200">
              <span className="h-2 w-2 rounded-full bg-blue-400" />
              CLOSED BETA · SHOPEE VN
            </p>
          </div>
        </div>

        {/* Right Side: Form */}
        <div className="flex flex-col justify-center px-6 py-10 sm:px-12 sm:py-14 lg:px-14 xl:px-16">
          <div className="w-full max-w-md lg:mx-auto">
            <Link href="/" className="mb-8 inline-flex items-center gap-1 text-sm font-bold text-slate-400 transition hover:text-slate-900 lg:hidden">
              ← Về trang chủ
            </Link>

            <h2 className="text-3xl font-black tracking-tight text-slate-950 sm:text-4xl">
              {isSignup ? "Tạo tài khoản" : "Chào mừng"}
            </h2>
            <p className="mt-3 mb-8 text-sm leading-relaxed text-slate-500 sm:text-base">
              {isSignup
                ? "Bắt đầu hành trình mua sắm thông minh ngay hôm nay."
                : "Đăng nhập để tiếp tục theo dõi các sản phẩm của bạn."}
            </p>

            {error && (
              <div className="mb-6 rounded-xl border border-red-200 bg-red-50 p-4">
                <p className="text-sm font-medium text-red-800">{error}</p>
              </div>
            )}

            {message && (
              <div className="mb-6 rounded-xl border border-emerald-200 bg-emerald-50 p-4">
                <p className="text-sm font-medium text-emerald-800">{message}</p>
              </div>
            )}

            <form onSubmit={handleLogin} className="space-y-5">
              <div>
                <label htmlFor="login-email" className="mb-2 block text-xs font-extrabold uppercase tracking-wide text-slate-500">
                  Địa chỉ Email
                </label>
                <input
                  id="login-email"
                  type="email"
                  className="w-full rounded-xl border border-slate-200 bg-white px-4 py-3.5 text-sm font-medium outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10"
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  placeholder="name@example.com"
                  required
                />
              </div>
              <div>
                <label htmlFor="login-password" className="mb-2 block text-xs font-extrabold uppercase tracking-wide text-slate-500">
                  Mật khẩu
                </label>
                <input
                  id="login-password"
                  type="password"
                  className="w-full rounded-xl border border-slate-200 bg-white px-4 py-3.5 text-sm font-medium outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  placeholder="••••••••"
                  required
                />
              </div>

              <button
                disabled={submitting}
                type="submit"
                className="mt-2 flex w-full items-center justify-center rounded-xl bg-slate-950 py-4 text-sm font-extrabold text-white shadow-lg shadow-slate-200 transition hover:-translate-y-0.5 hover:bg-blue-600 disabled:cursor-not-allowed disabled:bg-slate-400 disabled:hover:translate-y-0"
              >
                {submitting ? (
                  <span className="flex items-center gap-2">
                    <svg className="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" className="opacity-25"></circle><path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" className="opacity-75"></path></svg>
                    Đang xử lý...
                  </span>
                ) : (
                  isSignup ? "Tạo tài khoản" : "Đăng nhập"
                )}
              </button>
            </form>

            <div className="mt-8 text-center">
              <button
                type="button"
                onClick={() => { setIsSignup((current) => !current); setError(""); setMessage(""); }}
                className="text-sm font-bold text-slate-500 transition hover:text-slate-900"
              >
                {isSignup ? (
                  <span>Đã có tài khoản? <span className="text-blue-600">Đăng nhập</span></span>
                ) : (
                  <span>Chưa có tài khoản? <span className="text-blue-600">Tạo tài khoản</span></span>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
