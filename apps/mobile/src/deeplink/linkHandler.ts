export interface DeepLink {
  productId?: number;
  ref?: string;
}

export function parseDeepLink(url: string): DeepLink {
  try {
    const u = new URL(url);
    const productId = Number(u.searchParams.get("product_id") || "");
    const ref = u.searchParams.get("ref") || undefined;
    return {
      productId: Number.isFinite(productId) && productId > 0 ? productId : undefined,
      ref: ref || undefined,
    };
  } catch {
    return {};
  }
}
