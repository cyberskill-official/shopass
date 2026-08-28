"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { logout } from "@/lib/auth";

const links = [
  { label: "Bảng điều khiển", href: "/dashboard" },
  { label: "Cảnh báo", href: "/alerts" },
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
            <Link
              href="/dashboard"
              aria-label="Shopass — Bảng điều khiển"
              className="flex cursor-pointer items-center gap-2.5 transition hover:opacity-80 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/20"
            >
              <span className="grid h-9 w-9 place-items-center rounded-xl bg-gradient-to-br from-sky-500 to-teal-600 text-lg font-black text-white shadow-md shadow-sky-500/20">
                S
              </span>
              <span className="text-xl font-extrabold tracking-tight">
                Shop<span className="text-sky-600">ass</span>
              </span>
            </Link>

            {/* Desktop Navigation */}
            <nav className="hidden items-center gap-1 md:flex" aria-label="Chính">
              {links.map((link) => {
                const isActive = pathname?.startsWith(link.href);
                return (
                  <Link
                    key={link.href}
                    href={link.href}
                    className={`cursor-pointer rounded-xl px-4 py-2 text-sm font-bold transition focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/20 ${
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

          <div className="flex items-center gap-3 sm:gap-4">
            <div className="hidden items-center gap-2 rounded-full border border-sky-100 bg-sky-50 py-1.5 pl-2.5 pr-3 text-xs font-bold text-sky-800 shadow-sm sm:flex">
              <span className="relative flex h-2 w-2" aria-hidden="true">
                <span className="relative inline-flex h-2 w-2 rounded-full bg-sky-500" />
              </span>
              Closed beta
            </div>

            <button
              type="button"
              onClick={handleLogout}
              className="hidden cursor-pointer rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm font-bold text-slate-600 shadow-sm transition hover:bg-slate-50 hover:text-red-600 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/20 md:block"
            >
              Đăng xuất
            </button>

            <button
              type="button"
              className="cursor-pointer rounded-lg p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-900 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/20 md:hidden"
              aria-expanded={mobileMenuOpen}
              aria-controls="app-shell-mobile-nav"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            >
              <span className="sr-only">Mở menu</span>
              {mobileMenuOpen ? (
                <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
              ) : (
                <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" /></svg>
              )}
            </button>
          </div>
        </div>

        {mobileMenuOpen && (
          <div id="app-shell-mobile-nav" className="border-t border-slate-200 bg-white px-4 py-4 shadow-lg md:hidden">
            <nav className="flex flex-col gap-2" aria-label="Di động">
              {links.map((link) => {
                const isActive = pathname?.startsWith(link.href);
                return (
                  <Link
                    key={link.href}
                    href={link.href}
                    onClick={() => setMobileMenuOpen(false)}
                    className={`cursor-pointer rounded-xl px-4 py-3 text-sm font-bold transition focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/20 ${
                      isActive
                        ? "bg-slate-900 text-white"
                        : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                    }`}
                  >
                    {link.label}
                  </Link>
                );
              })}
              <p className="px-4 pt-1 text-xs font-medium text-slate-400">
                Danh sách quan tâm sẽ có sau closed beta — hiện mọi sản phẩm nằm trên Bảng điều khiển.
              </p>
              <div className="mt-2 border-t border-slate-100 pt-2">
                <button
                  type="button"
                  onClick={handleLogout}
                  className="w-full cursor-pointer rounded-xl px-4 py-3 text-left text-sm font-bold text-red-600 transition hover:bg-red-50 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-red-500/20"
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
