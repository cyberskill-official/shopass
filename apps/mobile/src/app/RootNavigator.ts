export type Screen = "Login" | "Home" | "Track" | "Checkout" | "Product";

export function resolveScreen(hasSession: boolean, deepLinkProductId?: number): Screen {
  if (!hasSession) return "Login";
  if (deepLinkProductId && deepLinkProductId > 0) return "Product";
  return "Home";
}
