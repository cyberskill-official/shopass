import type { MetadataRoute } from "next";
import { postsByDateDesc } from "@/lib/content/posts";
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
    { url: `${siteURL}/so-sanh/shopass-vs-beecost`, lastModified: new Date() },
    { url: `${siteURL}/thay-the-honey`, lastModified: new Date() },
    { url: `${siteURL}/blog`, lastModified: new Date() },
    { url: `${siteURL}/changelog`, lastModified: new Date() },
    ...postsByDateDesc().map((p) => ({
      url: `${siteURL}/blog/${p.slug}`,
      lastModified: new Date(p.date),
    })),
  ];
  return [...keywords, ...legal];
}
