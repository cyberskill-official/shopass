import type { MetadataRoute } from "next";
import { KEYWORD_PAGES } from "@/lib/seo/keywords";
import { siteURL } from "@/lib/site-url";

export default function sitemap(): MetadataRoute.Sitemap {
  const keywords = KEYWORD_PAGES.map((p) => ({
    url: `${siteURL}/${p.slug}`,
    lastModified: new Date(),
  }));
  const legal = [
    { url: `${siteURL}/chinh-sach-bao-mat`, lastModified: new Date() },
    { url: `${siteURL}/dieu-khoan`, lastModified: new Date() },
    { url: `${siteURL}/minh-bach`, lastModified: new Date() },
    { url: `${siteURL}/bang-gia`, lastModified: new Date() },
    { url: `${siteURL}/kiem-tra-sale-ao`, lastModified: new Date() },
    { url: `${siteURL}/lich-sale`, lastModified: new Date() },
  ];
  return [...keywords, ...legal];
}
