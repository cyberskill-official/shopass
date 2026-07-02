import { apiFetch } from "@/lib/api";

export interface Wishlist {
  id: number;
  name: string;
  itemCount: number;
}

export interface WishlistItem {
  id: number;
  productId: number;
  targetPrice: number | null; // VND int64
}

export async function createWishlist(name: string): Promise<Wishlist> {
  const res = await apiFetch("/v1/wishlists", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
  if (!res.ok) throw new Error("Không thể tạo wishlist");
  return (await res.json()) as Wishlist;
}

export async function listWishlists(): Promise<Wishlist[]> {
  const res = await apiFetch("/v1/wishlists");
  if (!res.ok) throw new Error("Không thể tải wishlist");
  return (await res.json()) as Wishlist[];
}

export async function deleteWishlist(id: number): Promise<void> {
  const res = await apiFetch(`/v1/wishlists/${id}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Không thể xóa wishlist");
}

export async function addItem(
  wishlistId: number,
  productId: number,
  targetPrice: number | null
): Promise<WishlistItem> {
  const res = await apiFetch(`/v1/wishlists/${wishlistId}/items`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ product_id: productId, target_price: targetPrice }), // int VND (DEC-WEB-19)
  });
  if (!res.ok) throw new Error("Không thể thêm sản phẩm vào wishlist");
  return (await res.json()) as WishlistItem;
}

export async function removeItem(wishlistId: number, itemId: number): Promise<void> {
  const res = await apiFetch(`/v1/wishlists/${wishlistId}/items/${itemId}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Không thể xóa sản phẩm khỏi wishlist");
}
