import { CHANGELOG } from "../lib/content/changelog";
import { BLOG_POSTS, postsByDateDesc } from "../lib/content/posts";
import { GET as rssGET } from "../app/rss.xml/route";

describe("R47 blog + changelog + RSS", () => {
  it("has two seed posts", () => {
    expect(BLOG_POSTS).toHaveLength(2);
    expect(BLOG_POSTS.some((p) => p.slug.includes("sale"))).toBe(true);
    expect(BLOG_POSTS.some((p) => p.slug.includes("chao"))).toBe(true);
  });

  it("changelog lists current release first", () => {
    expect(CHANGELOG[0]?.version).toBeTruthy();
    expect(CHANGELOG[0].items.length).toBeGreaterThan(0);
  });

  it("RSS is well-formed and lists posts", async () => {
    const res = await rssGET();
    const xml = await res.text();
    expect(res.headers.get("Content-Type")).toMatch(/rss\+xml/);
    expect(xml).toContain("<rss version=\"2.0\">");
    for (const p of postsByDateDesc()) {
      expect(xml).toContain(p.slug);
      expect(xml).toContain("<item>");
    }
  });
});
