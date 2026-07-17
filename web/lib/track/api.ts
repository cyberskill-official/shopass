import { apiFetch } from "@/lib/api";

export interface TrackResponse {
  product_id: number;
  platform: string;
  already_tracked: boolean;
}

export interface TrackedProduct {
  product_id: number;
  platform: string;
  platform_item_id: string;
  first_seen: string;
  tracked_at: string;
}

async function responseError(response: Response, fallback: string): Promise<Error> {
  try {
    const body: unknown = await response.json();
    if (
      typeof body === "object" &&
      body !== null &&
      "error" in body &&
      typeof body.error === "string" &&
      body.error.length > 0
    ) {
      return new Error(body.error);
    }
  } catch {
    // Use the safe, user-facing fallback when an upstream error has no JSON.
  }
  return new Error(fallback);
}

export async function trackShopeeProduct(itemURL: string): Promise<TrackResponse> {
  const response = await apiFetch("/v1/track", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ platform: "shopee", item_url: itemURL }),
  });
  if (!response.ok) {
    throw await responseError(response, "Không thể thêm sản phẩm để theo dõi.");
  }
  return (await response.json()) as TrackResponse;
}

export async function listTrackedProducts(): Promise<TrackedProduct[]> {
  const response = await apiFetch("/v1/tracked-products");
  if (!response.ok) {
    throw await responseError(response, "Không thể tải sản phẩm đang theo dõi.");
  }
  return (await response.json()) as TrackedProduct[];
}
