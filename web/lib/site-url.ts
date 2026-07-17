const configuredSiteURL = process.env.NEXT_PUBLIC_SITE_URL || "https://sandeal.vn";

export const siteURL = configuredSiteURL.replace(/\/$/, "");
