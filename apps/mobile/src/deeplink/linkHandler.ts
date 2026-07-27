export interface DeepLink {
  productId?: number;
  ref?: string;
}

export function parseDeepLink(url: string): DeepLink {
  try {
    const u = new URL(url);
    const rawId = u.searchParams.get("product_id") || "";
    const productId = Number(rawId);
    const ref = u.searchParams.get("ref") || undefined;
    return {
      productId: Number.isFinite(productId) && productId > 0 ? productId : undefined,
      ref: ref && /^[A-Za-z0-9_-]{2,32}$/.test(ref) ? ref : undefined,
    };
  } catch {
    return {};
  }
}

/** Invalid/missing product_id → Home; bad ref dropped but product still opens. */
export function routeFromDeepLink(link: DeepLink): "Home" | "Product" {
  return link.productId && link.productId > 0 ? "Product" : "Home";
}
