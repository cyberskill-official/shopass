import React from 'react';
import Link from 'next/link';

export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      <header className="bg-white shadow-sm border-b px-6 py-4 flex justify-between items-center">
        <h1 className="text-xl font-bold text-blue-600">SănDeal</h1>
        <nav className="flex space-x-4">
          <Link href="/dashboard" className="block px-4 py-2 hover:bg-gray-100 rounded text-gray-700">
            Dashboard
          </Link>
          <Link href="/wishlist" className="block px-4 py-2 hover:bg-gray-100 rounded text-gray-700">
            Wishlist
          </Link>
          <Link href="/alerts" className="block px-4 py-2 hover:bg-gray-100 rounded text-gray-700">
            Cảnh báo
          </Link>
          <Link href="/products" className="block px-4 py-2 hover:bg-gray-100 rounded text-gray-700">
            Sản phẩm
          </Link>
        </nav>
      </header>
      <main className="flex-1 p-6">
        {children}
      </main>
    </div>
  );
}
