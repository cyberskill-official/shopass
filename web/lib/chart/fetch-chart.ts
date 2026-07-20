import { apiFetch } from "@/lib/api";
import { RANGE_ALLOWLIST, type ChartResponse, type Range } from "./types";

export type BrowserSnapshotResponse = {
  written: boolean;
};

class NotFoundError extends Error {
  constructor(message = "Not found") {
    super(message);
    this.name = "NotFoundError";
  }
}

export async function fetchChart(productId: number, range: Range): Promise<ChartResponse> {
  if (!RANGE_ALLOWLIST.includes(range)) {
    throw new Error("range ngoài allowlist");
  }
  
  const res = await apiFetch(`/v1/products/${productId}/chart?range=${range}`);
  
  if (res.status === 404) {
    throw new NotFoundError();
  }
  if (!res.ok) {
    throw new Error("không tải được biểu đồ");
  }
  
  return (await res.json()) as ChartResponse;
}

/**
 * Stores a price that the signed-in user has explicitly checked in their own
 * browser. This is intentionally separate from Shopass's automated collection
 * flow: no marketplace session or browser data is sent to the API.
 */
export async function submitBrowserPrice(
  productId: number,
  price: number
): Promise<BrowserSnapshotResponse> {
  if (!Number.isSafeInteger(productId) || productId <= 0) {
    throw new Error("Sản phẩm không hợp lệ");
  }

  if (!Number.isSafeInteger(price) || price <= 0 || price > 1_000_000_000_000) {
    throw new Error("Nhập giá hợp lệ bằng VNĐ");
  }

  const res = await apiFetch(`/v1/products/${productId}/browser-snapshot`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ price }),
  });

  if (!res.ok) {
    throw new Error("Không thể lưu giá lúc này. Vui lòng thử lại.");
  }

  return (await res.json()) as BrowserSnapshotResponse;
}
