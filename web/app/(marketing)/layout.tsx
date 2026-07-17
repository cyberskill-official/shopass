import Link from "next/link";

export default function MarketingLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="min-h-screen flex flex-col bg-white">
      <header className="bg-blue-600 text-white shadow-sm px-6 py-4 flex justify-between items-center">
        <h1 className="text-xl font-bold">SănDeal</h1>
        <nav>
          <Link href="/login" className="hover:underline">Đăng nhập</Link>
        </nav>
      </header>
      <main className="flex-1">
        {children}
      </main>
      <footer className="bg-gray-100 py-6 text-center text-gray-600">
        <p>&copy; 2026 SănDeal. All rights reserved.</p>
      </footer>
    </div>
  );
}
