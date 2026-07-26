import Link from "next/link";

export default function MarketingLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="min-h-screen flex flex-col bg-white">
      <header className="bg-blue-600 text-white shadow-sm px-6 py-4 flex justify-between items-center">
        <h1 className="text-xl font-bold">Shopass</h1>
        <nav>
          <Link href="/login" className="hover:underline">Đăng nhập</Link>
        </nav>
      </header>
      <main className="flex-1">
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
