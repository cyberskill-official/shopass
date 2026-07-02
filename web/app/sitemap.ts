import type { MetadataRoute } from "next";
import { KEYWORD_PAGES } from "@/lib/seo/keywords";

export default function sitemap(): MetadataRoute.Sitemap {
  return KEYWORD_PAGES.map((p) => ({
    url: `https://sandeal.vn/${p.slug}`,
    lastModified: new Date(),
  }));
}
