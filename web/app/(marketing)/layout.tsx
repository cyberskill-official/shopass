import Link from "next/link";

export default function MarketingLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="flex min-h-screen flex-col bg-[#f4f7fb] text-slate-900">
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-white focus:px-3 focus:py-2 focus:text-sm focus:font-bold focus:text-slate-900 focus:shadow-lg"
      >
        Bỏ qua điều hướng
      </a>
      <header className="sticky top-0 z-40 border-b border-slate-200/60 bg-white/80 backdrop-blur-xl">
        <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4 sm:px-6">
          <Link
            href="/"
            className="flex items-center gap-2 text-lg font-black tracking-tight"
            aria-label="Shopass — Trang chủ"
          >
            <span className="grid h-8 w-8 place-items-center rounded-lg bg-gradient-to-br from-sky-500 to-teal-600 text-sm font-black text-white shadow-sm shadow-sky-500/20">
              S
            </span>
            Shop<span className="text-sky-600">ass</span>
          </Link>
          <nav aria-label="Tài khoản" className="flex items-center gap-4">
            <Link
              href="/login"
              className="text-sm font-bold text-slate-600 transition hover:text-sky-800 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-sky-500/20"
            >
              Đăng nhập
            </Link>
          </nav>
        </div>
      </header>
      <main id="main" tabIndex={-1} className="flex-1 outline-none">
        {children}
      </main>
      <footer className="border-t border-slate-200/70 bg-white/60 px-6 py-8 text-center text-sm text-slate-500">
        <p className="flex flex-wrap items-center justify-center gap-x-4 gap-y-2">
          <Link href="/chinh-sach-bao-mat" className="font-medium hover:text-sky-800 hover:underline">
            Chính sách bảo mật
          </Link>
          <Link href="/dieu-khoan" className="font-medium hover:text-sky-800 hover:underline">
            Điều khoản
          </Link>
          <Link href="/minh-bach" className="font-medium hover:text-sky-800 hover:underline">
            Minh bạch
          </Link>
          <a href="mailto:info@cyberskill.world" className="font-medium hover:text-sky-800 hover:underline">
            DSAR
          </a>
        </p>
        <p className="mt-3 text-slate-400">&copy; 2026 Shopass · CyberSkill JSC</p>
      </footer>
    </div>
  );
}
