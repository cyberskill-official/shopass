import { apiFetch } from "@/lib/api";
import { RANGE_ALLOWLIST, type ChartResponse, type Range } from "./types";

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
