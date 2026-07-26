import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { BLOG_POSTS, getPost } from "@/lib/content/posts";
import { siteURL } from "@/lib/site-url";

type Props = { params: Promise<{ slug: string }> };

export function generateStaticParams() {
  return BLOG_POSTS.map((p) => ({ slug: p.slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const post = getPost(slug);
  if (!post) return {};
  const url = `${siteURL}/blog/${post.slug}`;
  return {
    title: post.title,
    description: post.description,
    alternates: { canonical: url },
    openGraph: { title: post.title, description: post.description, url, type: "article", locale: "vi_VN" },
  };
}

export default async function BlogPostPage({ params }: Props) {
  const { slug } = await params;
  const post = getPost(slug);
  if (!post) notFound();

  const url = `${siteURL}/blog/${post.slug}`;
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Article",
    headline: post.title,
    description: post.description,
    datePublished: post.date,
    dateModified: post.date,
    inLanguage: "vi-VN",
    author: { "@type": "Organization", name: "Shopass" },
    mainEntityOfPage: url,
  };

  return (
    <article className="mx-auto max-w-3xl px-6 py-12">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
      <p className="text-sm font-bold text-sky-800">
        <Link href="/blog">← Blog</Link>
      </p>
      <p className="mt-4 text-xs font-bold uppercase tracking-wide text-slate-500">{post.date}</p>
      <h1 className="mt-2 text-3xl font-black tracking-tight text-slate-900">{post.title}</h1>
      <p className="mt-3 text-slate-600">{post.description}</p>
      <div className="mt-8 space-y-4 text-sm leading-relaxed text-slate-800">
        {post.body.map((para) => (
          <p key={para.slice(0, 48)}>{para}</p>
        ))}
      </div>
      <p className="mt-10 text-sm">
        <Link href="/kiem-tra-sale-ao" className="font-bold text-sky-800 underline">
          Thử kiểm tra sale ảo
        </Link>
        {" · "}
        <Link href="/login?signup=1&next=/onboarding" className="font-bold text-sky-800 underline">
          Tạo tài khoản
        </Link>
      </p>
    </article>
  );
}
