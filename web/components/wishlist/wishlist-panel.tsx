import React, { useState, useEffect } from "react";
import { WishlistItemRow } from "./wishlist-item-row";
import {
  listWishlists,
  createWishlist,
  deleteWishlist,
  addItem,
  removeItem,
  type Wishlist,
  type WishlistItem,
} from "@/lib/wishlist/api";
import { apiFetch } from "@/lib/api";

export function WishlistPanel() {
  const [wishlists, setWishlists] = useState<Wishlist[]>([]);
  const [items, setItems] = useState<WishlistItem[]>([]); // for simplicity, assuming a single list for MVP
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // New wishlist state
  const [newListName, setNewListName] = useState("");

  // New item state
  const [newProductId, setNewProductId] = useState("");
  const [newTargetPrice, setNewTargetPrice] = useState("");

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const lists = await listWishlists();
      setWishlists(lists);

      if (lists.length > 0) {
        // Fetch items for the first list as a simplified UX for this slice
        const res = await apiFetch(`/v1/wishlists/${lists[0].id}/items`);
        if (res.ok) {
          const fetchedItems = await res.json();
          setItems(fetchedItems);
        } else if (res.status === 403 || res.status === 404) {
          setError("Không tìm thấy danh sách");
        }
      }
    } catch (error) {
      setError(
        error instanceof Error && error.message
          ? error.message
          : "Đã xảy ra lỗi khi tải danh sách"
      );
    } finally {
      setLoading(false);
    }
  };

  const handleCreateList = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newListName.trim()) return;
    try {
      await createWishlist(newListName);
      setNewListName("");
      await loadData();
    } catch (error) {
      alert(error instanceof Error && error.message ? error.message : "Đã xảy ra lỗi");
    }
  };

  const handleAddItem = async (e: React.FormEvent) => {
    e.preventDefault();
    if (wishlists.length === 0 || !newProductId.trim()) return;
    try {
      const pid = parseInt(newProductId, 10);
      const tp = newTargetPrice.trim() === "" ? null : parseInt(newTargetPrice, 10);

      if (tp !== null && (isNaN(tp) || tp <= 0)) {
        alert("Giá phải là số nguyên dương");
        return;
      }

      await addItem(wishlists[0].id, pid, tp);
      setNewProductId("");
      setNewTargetPrice("");
      await loadData();
    } catch (error) {
      alert(error instanceof Error && error.message ? error.message : "Đã xảy ra lỗi");
    }
  };

  const handleDeleteList = async (id: number) => {
    if (!confirm("Bạn có chắc chắn muốn xóa wishlist này?")) return;
    try {
      await deleteWishlist(id);
      await loadData();
    } catch (error) {
      alert(error instanceof Error && error.message ? error.message : "Đã xảy ra lỗi");
    }
  };

  if (loading && wishlists.length === 0) {
    return <div className="text-gray-500">Đang tải...</div>;
  }

  if (error) {
    return <div className="text-red-500">{error}</div>;
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <h2 className="text-xl font-semibold mb-4">Quản lý Wishlist</h2>

      {wishlists.length === 0 ? (
        <div className="bg-gray-50 p-6 text-center rounded-lg border border-dashed border-gray-300">
          <p className="text-gray-500 mb-4">Bạn chưa có wishlist nào.</p>
          <form onSubmit={handleCreateList} className="flex justify-center items-center space-x-2">
            <input
              type="text"
              value={newListName}
              onChange={(e) => setNewListName(e.target.value)}
              placeholder="Tên wishlist mới"
              className="border border-gray-300 rounded px-3 py-2 text-sm"
            />
            <button type="submit" className="bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700">
              Tạo Wishlist
            </button>
          </form>
        </div>
      ) : (
        <div>
          <div className="flex justify-between items-center mb-6 bg-gray-50 p-4 rounded-lg">
            <div>
              <span className="font-medium text-gray-800">{wishlists[0].name}</span>
              <span className="text-gray-500 text-sm ml-2">({wishlists[0].itemCount} sản phẩm)</span>
            </div>
            <button
              onClick={() => handleDeleteList(wishlists[0].id)}
              className="text-red-600 text-sm hover:underline"
            >
              Xóa danh sách
            </button>
          </div>

          <form onSubmit={handleAddItem} className="flex space-x-2 mb-6 items-end">
            <div>
              <label className="block text-xs text-gray-500 mb-1">ID Sản phẩm</label>
              <input
                type="number"
                required
                value={newProductId}
                onChange={(e) => setNewProductId(e.target.value)}
                placeholder="Ví dụ: 90112"
                className="border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="block text-xs text-gray-500 mb-1">Giá mong muốn (tùy chọn)</label>
              <input
                type="number"
                step={1}
                value={newTargetPrice}
                onChange={(e) => setNewTargetPrice(e.target.value)}
                placeholder="Ví dụ: 89000"
                className="border border-gray-300 rounded px-3 py-2 text-sm"
              />
            </div>
            <button type="submit" className="bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700">
              Thêm
            </button>
          </form>

          <div className="border border-gray-200 rounded-lg overflow-hidden">
            {items.length === 0 ? (
              <div className="p-4 text-center text-gray-500">Chưa có sản phẩm nào</div>
            ) : (
              items.map((item) => (
                <WishlistItemRow
                  key={item.id}
                  item={item}
                  onUpdateTargetPrice={async (id, newTargetPrice) => {
                    // Update by removing and adding, or patching if API supported it
                    // Assuming adding the same product overwrites or we have a patch
                    // Fallback to calling a patch if needed, or simply for UX updating state
                    // We'll update the item via API:
                    const res = await apiFetch(`/v1/wishlists/${wishlists[0].id}/items/${id}`, {
                      method: "PATCH",
                      headers: { "Content-Type": "application/json" },
                      body: JSON.stringify({ target_price: newTargetPrice })
                    });
                    if (!res.ok) throw new Error("Lỗi cập nhật");
                    await loadData();
                  }}
                  onRemove={async (id) => {
                    await removeItem(wishlists[0].id, id);
                    await loadData();
                  }}
                />
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
