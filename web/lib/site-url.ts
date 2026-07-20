const PRODUCTION_SITE_ORIGIN = "https://shopass.cyberskill.world";

/**
 * A bookmarklet executes in the marketplace's origin, so the destination
 * must never be an arbitrary deployment setting. Keep the production origin
 * explicit and allow localhost only for a local development build.
 */
export function trustedSiteURL(candidate: string | undefined): string {
  if (!candidate) return PRODUCTION_SITE_ORIGIN;

  try {
    const parsed = new URL(candidate);
    const isBareOrigin =
      parsed.pathname === "/" &&
      !parsed.search &&
      !parsed.hash &&
      !parsed.username &&
      !parsed.password;

    if (
      parsed.protocol === "https:" &&
      parsed.hostname === "shopass.cyberskill.world" &&
      !parsed.port &&
      isBareOrigin
    ) {
      return PRODUCTION_SITE_ORIGIN;
    }

    const isLocalDevelopmentOrigin =
      process.env.NODE_ENV !== "production" &&
      parsed.protocol === "http:" &&
      !parsed.port &&
      isBareOrigin &&
      (parsed.hostname === "localhost" || parsed.hostname === "127.0.0.1");
    if (isLocalDevelopmentOrigin) return parsed.origin;
  } catch {
    // Fall through to the production origin. Public configuration must never
    // redirect a marketplace bookmarklet to an arbitrary host.
  }

  return PRODUCTION_SITE_ORIGIN;
}

export const siteURL = trustedSiteURL(process.env.NEXT_PUBLIC_SITE_URL);
