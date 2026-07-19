"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { logout } from "@/lib/auth";

const links = [
  { label: "Bảng điều khiển", href: "/dashboard" },
];

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const handleLogout = async () => {
    try {
      await logout();
    } catch {
      // ignore
    }
    window.location.href = "/login";
  };

  return (
    <div className="mesh-bg min-h-screen text-slate-900">
      <header className="sticky top-0 z-40 border-b border-slate-200/60 bg-white/70 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">

          <div className="flex items-center gap-8">
            <Link href="/dashboard" className="flex items-center gap-2.5 transition hover:opacity-80">
              <span className="grid h-9 w-9 place-items-center rounded-xl bg-gradient-to-br from-blue-600 via-indigo-600 to-violet-600 text-lg font-black text-white shadow-md shadow-blue-500/20">S</span>
              <span className="text-xl font-extrabold tracking-tight">Shop<span className="text-blue-600">ass</span></span>
            </Link>

            {/* Desktop Navigation */}
            <nav className="hidden items-center gap-1 md:flex">
              {links.map((link) => {
                const isActive = pathname?.startsWith(link.href);
                return (
                  <Link
                    key={link.href}
                    href={link.href}
                    className={`rounded-xl px-4 py-2 text-sm font-bold transition ${
                      isActive
                        ? "bg-slate-900 text-white shadow-sm"
                        : "text-slate-600 hover:bg-slate-100 hover:text-slate-900"
                    }`}
                  >
                    {link.label}
                  </Link>
                );
              })}
            </nav>
          </div>

          <div className="flex items-center gap-4">
            <div className="hidden items-center gap-2 rounded-full border border-emerald-100 bg-emerald-50 py-1.5 pl-2.5 pr-3 text-xs font-bold text-emerald-700 shadow-sm sm:flex">
              <span className="relative flex h-2 w-2">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"></span>
                <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-500"></span>
              </span>
              Hệ thống ổn định
            </div>

            {/* User Dropdown / Logout placeholder */}
            <button onClick={handleLogout} className="hidden rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm font-bold text-slate-600 shadow-sm transition hover:bg-slate-50 hover:text-red-600 md:block">
              Đăng xuất
            </button>

            {/* Mobile menu button */}
            <button
              type="button"
              className="rounded-lg p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-900 md:hidden"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            >
              <span className="sr-only">Mở menu</span>
              {mobileMenuOpen ? (
                <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
              ) : (
                <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" /></svg>
              )}
            </button>
          </div>
        </div>

        {/* Mobile Navigation */}
        {mobileMenuOpen && (
          <div className="border-t border-slate-200 bg-white px-4 py-4 md:hidden shadow-lg">
            <nav className="flex flex-col gap-2">
              {links.map((link) => {
                const isActive = pathname?.startsWith(link.href);
                return (
                  <Link
                    key={link.href}
                    href={link.href}
                    onClick={() => setMobileMenuOpen(false)}
                    className={`rounded-xl px-4 py-3 text-sm font-bold transition ${
                      isActive
                        ? "bg-slate-900 text-white"
                        : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                    }`}
                  >
                    {link.label}
                  </Link>
                );
              })}
              <div className="mt-2 border-t border-slate-100 pt-2">
                <button
                  onClick={handleLogout}
                  className="w-full rounded-xl px-4 py-3 text-left text-sm font-bold text-red-600 transition hover:bg-red-50"
                >
                  Đăng xuất
                </button>
              </div>
            </nav>
          </div>
        )}
      </header>

      <main className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8 lg:py-10">
        {children}
      </main>
    </div>
  );
}
