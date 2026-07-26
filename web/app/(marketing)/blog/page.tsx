import type { Metadata } from "next";
import Link from "next/link";
import { postsByDateDesc } from "@/lib/content/posts";
import { siteURL } from "@/lib/site-url";

export const metadata: Metadata = {
  title: "Blog Shopass",
  description: "Hướng dẫn săn deal, sale ảo, và cập nhật sản phẩm.",
  alternates: { canonical: `${siteURL}/blog` },
};

export default function BlogIndexPage() {
  const posts = postsByDateDesc();
  return (
    <div className="mx-auto max-w-3xl px-6 py-12">
      <p className="text-sm font-bold text-sky-800">
        <Link href="/">Shopass</Link>
        {" · "}
        <Link href="/changelog">Changelog</Link>
        {" · "}
        <a href="/rss.xml">RSS</a>
      </p>
      <h1 className="mt-4 text-3xl font-black tracking-tight">Blog</h1>
      <p className="mt-2 text-slate-600">Ghi chú săn deal và câu chuyện sản phẩm — ngắn, kiểm chứng được.</p>
      <ul className="mt-10 space-y-8">
        {posts.map((p) => (
          <li key={p.slug}>
            <p className="text-xs font-bold uppercase tracking-wide text-slate-500">{p.date}</p>
            <Link href={`/blog/${p.slug}`} className="mt-1 block text-xl font-black text-slate-900 hover:text-sky-800">
              {p.title}
            </Link>
            <p className="mt-2 text-sm text-slate-600">{p.description}</p>
          </li>
        ))}
      </ul>
    </div>
  );
}
