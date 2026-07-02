import React, { useState } from "react";
import type { WishlistItem } from "@/lib/wishlist/api";

export function WishlistItemRow({
  item,
  onUpdateTargetPrice,
  onRemove,
}: {
  item: WishlistItem;
  onUpdateTargetPrice: (id: number, newTargetPrice: number | null) => Promise<void>;
  onRemove: (id: number) => Promise<void>;
}) {
  const [editing, setEditing] = useState(false);
  const [targetPrice, setTargetPrice] = useState<string>(
    item.targetPrice ? item.targetPrice.toString() : ""
  );
  const [loading, setLoading] = useState(false);

  const handleSave = async () => {
    setLoading(true);
    try {
      const parsed = targetPrice.trim() === "" ? null : parseInt(targetPrice, 10);
      if (parsed !== null && (isNaN(parsed) || parsed <= 0)) {
        alert("Giá mong muốn phải là số nguyên dương");
        return;
      }
      await onUpdateTargetPrice(item.id, parsed);
      setEditing(false);
    } catch (e: any) {
      alert(e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-between p-3 border-b border-gray-100 bg-white">
      <div>
        <span className="font-medium text-gray-800">Sản phẩm #{item.productId}</span>
      </div>
      <div className="flex items-center space-x-4">
        {editing ? (
          <div className="flex space-x-2">
            <input
              type="number"
              step={1}
              value={targetPrice}
              onChange={(e) => setTargetPrice(e.target.value)}
              placeholder="Giá mong muốn"
              className="border border-gray-300 rounded px-2 py-1 text-sm w-32"
              disabled={loading}
            />
            <button
              onClick={handleSave}
              disabled={loading}
              className="text-blue-600 hover:text-blue-800 text-sm font-medium disabled:opacity-50"
            >
              Lưu
            </button>
            <button
              onClick={() => setEditing(false)}
              disabled={loading}
              className="text-gray-500 hover:text-gray-700 text-sm disabled:opacity-50"
            >
              Hủy
            </button>
          </div>
        ) : (
          <div className="flex space-x-2 items-center">
            <span className="text-gray-600 text-sm">
              {item.targetPrice
                ? new Intl.NumberFormat("vi-VN").format(item.targetPrice) + " ₫"
                : "Chưa đặt giá mong muốn"}
            </span>
            <button
              onClick={() => setEditing(true)}
              className="text-blue-600 hover:text-blue-800 text-sm"
            >
              Sửa
            </button>
          </div>
        )}
        <button
          onClick={() => {
            if (confirm("Xóa sản phẩm này?")) {
              onRemove(item.id);
            }
          }}
          className="text-red-600 hover:text-red-800 text-sm"
        >
          Xóa
        </button>
      </div>
    </div>
  );
}
