import type { MetadataRoute } from "next";
import { KEYWORD_PAGES } from "@/lib/seo/keywords";
import { siteURL } from "@/lib/site-url";

export default function sitemap(): MetadataRoute.Sitemap {
  return KEYWORD_PAGES.map((p) => ({
    url: `${siteURL}/${p.slug}`,
    lastModified: new Date(),
  }));
}
