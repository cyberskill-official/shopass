import { postsByDateDesc } from "@/lib/content/posts";
import { siteURL } from "@/lib/site-url";

function escapeXml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&apos;");
}

export function buildRssXml(): string {
  const posts = postsByDateDesc();
  const items = posts
    .map((p) => {
      const link = `${siteURL}/blog/${p.slug}`;
      return `    <item>
      <title>${escapeXml(p.title)}</title>
      <link>${escapeXml(link)}</link>
      <guid>${escapeXml(link)}</guid>
      <pubDate>${new Date(p.date + "T00:00:00Z").toUTCString()}</pubDate>
      <description>${escapeXml(p.description)}</description>
    </item>`;
    })
    .join("\n");

  return `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Shopass Blog</title>
    <link>${escapeXml(siteURL)}/blog</link>
    <description>Hướng dẫn săn deal và cập nhật Shopass</description>
    <language>vi-VN</language>
${items}
  </channel>
</rss>
`;
}
