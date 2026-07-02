"use client";

import React from "react";
import { WishlistPanel } from "@/components/wishlist/wishlist-panel";

export default function WishlistPage() {
  return (
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-2xl font-bold text-gray-800 mb-6">Wishlist của tôi</h1>
      <WishlistPanel />
    </div>
  );
}
