/** Auth-gated navigation shell (TASK-MOBILE-001). */

export type Screen = "Login" | "Home" | "Track" | "Checkout" | "Product";

export type Tab = "Home" | "Track" | "Checkout";

export function resolveScreen(hasSession: boolean, deepLinkProductId?: number): Screen {
  if (!hasSession) return "Login";
  if (deepLinkProductId && deepLinkProductId > 0) return "Product";
  return "Home";
}

export function resolveTab(screen: Screen): Tab | null {
  switch (screen) {
    case "Home":
      return "Home";
    case "Track":
      return "Track";
    case "Checkout":
      return "Checkout";
    case "Login":
    case "Product":
      return null;
    default: {
      const _exhaustive: never = screen;
      return _exhaustive;
    }
  }
}

export const APP_TABS: Tab[] = ["Home", "Track", "Checkout"];
