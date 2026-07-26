import Link from "next/link";

export default function MarketingLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="flex min-h-screen flex-col bg-white">
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-white focus:px-3 focus:py-2 focus:text-sm focus:font-bold focus:text-slate-900"
      >
        Bỏ qua điều hướng
      </a>
      <header className="flex items-center justify-between bg-blue-600 px-6 py-4 text-white shadow-sm">
        <Link href="/" className="text-xl font-bold">
          Shopass
        </Link>
        <nav aria-label="Tài khoản">
          <Link href="/login" className="hover:underline">
            Đăng nhập
          </Link>
        </nav>
      </header>
      <main id="main" className="flex-1">
        {children}
      </main>
      <footer className="bg-gray-100 px-6 py-6 text-center text-sm text-gray-600">
        <p className="flex flex-wrap items-center justify-center gap-x-4 gap-y-2">
          <Link href="/chinh-sach-bao-mat" className="hover:underline">
            Chính sách bảo mật
          </Link>
          <Link href="/dieu-khoan" className="hover:underline">
            Điều khoản
          </Link>
          <Link href="/minh-bach" className="hover:underline">
            Minh bạch
          </Link>
          <a href="mailto:info@cyberskill.world" className="hover:underline">
            DSAR
          </a>
        </p>
        <p className="mt-3">&copy; 2026 Shopass · CyberSkill JSC</p>
      </footer>
    </div>
  );
}
