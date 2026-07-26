import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { KEYWORD_PAGES, relatedPages } from "@/lib/seo/keywords";
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
  const related = relatedPages(p);
  const jsonLd = buildJsonLd(p, url);
  const scripts = Array.isArray(jsonLd) ? jsonLd : [jsonLd];

  return (
    <article lang="vi-VN" className="mx-auto max-w-3xl px-6 py-12 text-slate-900">
      {scripts.map((block, i) => (
        <script
          // eslint-disable-next-line react/no-danger
          key={i}
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(block) }}
        />
      ))}

      <p className="text-sm font-bold text-sky-800">
        <Link href="/">Shopass</Link>
        {" · "}
        <Link href="/blog">Blog</Link>
        {p.cluster === "calendar" && (
          <>
            {" · "}
            <Link href="/lich-sale">Lịch sale</Link>
          </>
        )}
        {p.cluster === "verdict" && (
          <>
            {" · "}
            <Link href="/kiem-tra-sale-ao">Kiểm tra sale ảo</Link>
          </>
        )}
      </p>

      <h1 className="mt-4 text-3xl font-black tracking-tight sm:text-4xl">{p.keyword}</h1>
      <p className="mt-3 text-slate-600">{p.description}</p>

      <div className="mt-8 space-y-4 text-sm leading-relaxed text-slate-800">
        {p.intro.map((para) => (
          <p key={para.slice(0, 40)}>{para}</p>
        ))}
      </div>

      {p.saleDate && (
        <div className="mt-8 border border-sky-100 bg-sky-50/80 px-5 py-4 text-sm">
          <p className="font-bold text-sky-950">Mốc sale: {p.saleDate}</p>
          <p className="mt-1 text-sky-900">
            Đặt nhắc hoặc xem đếm ngược tại{" "}
            <Link href="/lich-sale" className="font-bold underline">
              công cụ lịch sale
            </Link>
            .
          </p>
        </div>
      )}

      {p.faqs.length > 0 && (
        <section className="mt-10">
          <h2 className="text-xl font-black">Câu hỏi thường gặp</h2>
          <div className="mt-4 divide-y divide-slate-200">
            {p.faqs.map((f) => (
              <details key={f.q} className="py-3">
                <summary className="cursor-pointer font-bold text-slate-900">{f.q}</summary>
                <p className="mt-2 text-sm text-slate-600">{f.a}</p>
              </details>
            ))}
          </div>
        </section>
      )}

      <section className="mt-10">
        <h2 className="text-xl font-black">Liên quan</h2>
        <ul className="mt-3 flex flex-wrap gap-2">
          {related.map((r) => (
            <li key={r.href}>
              <Link
                href={r.href}
                className="inline-flex rounded-full border border-slate-200 px-3 py-1.5 text-xs font-bold text-slate-700 hover:border-sky-300 hover:text-sky-900"
              >
                {r.label}
              </Link>
            </li>
          ))}
        </ul>
      </section>

      <div className="mt-10 border border-slate-200 bg-slate-50 px-6 py-6 text-center">
        <h2 className="text-lg font-black text-slate-900">Sẵn sàng săn sale bằng dữ liệu?</h2>
        <p className="mt-2 text-sm text-slate-600">Theo dõi giá và cảnh báo đáy — miễn phí để bắt đầu.</p>
        <Link
          href="/login?signup=1&next=/onboarding"
          className="mt-4 inline-block rounded-xl bg-slate-950 px-6 py-3 text-sm font-extrabold text-white"
        >
          Dùng Shopass miễn phí
        </Link>
      </div>
    </article>
  );
}
