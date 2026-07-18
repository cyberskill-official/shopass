import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { KEYWORD_PAGES } from "@/lib/seo/keywords";
import { buildJsonLd } from "@/lib/seo/jsonld";
import { siteURL } from "@/lib/site-url";

export function generateStaticParams() {
  return KEYWORD_PAGES.map((p) => ({ keyword: p.slug }));
}

type KeywordPageProps = { params: Promise<{ keyword: string }> };

export async function generateMetadata({ params }: KeywordPageProps): Promise<Metadata> {
  const { keyword } = await params;
  const p = KEYWORD_PAGES.find((k) => k.slug === keyword);
  if (!p) return {};

  const url = `${siteURL}/${p.slug}`;
  return {
    title: p.title,
    description: p.description,
    alternates: { canonical: url },
    openGraph: {
      title: p.title,
      description: p.description,
      url,
      type: "article",
      locale: "vi_VN",
    },
    twitter: {
      card: "summary_large_image",
      title: p.title,
      description: p.description,
    },
  };
}

export default async function KeywordPage({ params }: KeywordPageProps) {
  const { keyword } = await params;
  const p = KEYWORD_PAGES.find((k) => k.slug === keyword);
  if (!p) {
    notFound();
  }

  const url = `${siteURL}/${p.slug}`;

  return (
    <article lang="vi-VN" className="max-w-3xl mx-auto px-6 py-12">
      <h1 className="text-3xl font-bold mb-6">{p.keyword}</h1>
      <div className="prose mb-8">
        <p>{p.description}</p>
        <p>Shopass là công cụ cảnh báo giá và kiểm tra sale ảo số 1 Việt Nam. Đừng để bị lừa bởi những chiêu trò tăng giá rồi giảm ảo.</p>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 my-8 text-center">
        <h2 className="text-xl font-semibold text-blue-800 mb-4">Bạn đã sẵn sàng săn sale thật chưa?</h2>
        <Link
          href="/login?signup=1"
          className="inline-block bg-blue-600 text-white font-medium py-3 px-8 rounded-full hover:bg-blue-700 transition"
        >
          Dùng Shopass miễn phí
        </Link>
      </div>

      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(buildJsonLd(p, url)) }}
      />
    </article>
  );
}
