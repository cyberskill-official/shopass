import React from 'react';
import Link from 'next/link';

export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      <header className="bg-white shadow-sm border-b px-6 py-4 flex justify-between items-center">
        <h1 className="text-xl font-bold text-blue-600">SănDeal</h1>
        <nav className="flex space-x-4">
          <Link href="/dashboard" className="text-gray-600 hover:text-blue-600">Dashboard</Link>
          <Link href="/wishlist" className="text-gray-600 hover:text-blue-600">Wishlist</Link>
          <Link href="/alerts" className="text-gray-600 hover:text-blue-600">Alerts</Link>
        </nav>
      </header>
      <main className="flex-1 p-6">
        {children}
      </main>
    </div>
  );
}
