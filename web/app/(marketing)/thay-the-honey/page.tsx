import type { Metadata } from "next";
import Link from "next/link";
import { SignupCta } from "@/components/landing/landing-cta";
import { siteURL } from "@/lib/site-url";

const title = "Thay thế Honey tại Việt Nam | Shopass";
const description =
  "Honey bị cắt affiliate vì ghi đè attribution ẩn. Shopass chọn kiến trúc đối lập: bạn thấy dữ liệu nào rời máy, affiliate chỉ khi bạn bấm.";

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: `${siteURL}/thay-the-honey` },
  openGraph: {
    title,
    description,
    url: `${siteURL}/thay-the-honey`,
    locale: "vi_VN",
    type: "article",
  },
};

const SOURCES = [
  {
    label: "Rakuten Advertising — Honey removed (reporting)",
    href: "https://www.theverge.com/2024/12/23/24328767/honey-paypal-rakuten-affiliate-marketing",
  },
  {
    label: "Impact.com — Honey removed from network (reporting)",
    href: "https://www.theverge.com/2024/12/30/24332751/honey-paypal-impact-affiliate-network",
  },
];

const jsonLd = {
  "@context": "https://schema.org",
  "@type": "WebPage",
  name: title,
  description,
  url: `${siteURL}/thay-the-honey`,
  inLanguage: "vi-VN",
  datePublished: "2026-07-26",
  dateModified: "2026-07-26",
  citation: SOURCES.map((s) => s.href),
};

export default function HoneyAlternativePage() {
  return (
    <article className="mx-auto max-w-3xl px-6 py-12 text-slate-900">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
      <p className="text-sm font-bold text-sky-800">
        <Link href="/">Shopass</Link>
        {" · "}
        <Link href="/minh-bach">Minh bạch</Link>
        {" · "}
        <Link href="/so-sanh/shopass-vs-beecost">vs BeeCost</Link>
      </p>
      <h1 className="mt-4 text-3xl font-black tracking-tight sm:text-4xl">{title}</h1>
      <p className="mt-3 text-slate-600">
        Người dùng Việt Nam cần công cụ săn deal không đánh cắp attribution. Shopass không phải clone
        Honey — và cố tình không copy mô hình cookie-stuffing.
      </p>

      <h2 className="mt-10 text-xl font-black">Vì sao Honey mất uy tín</h2>
      <p className="mt-3 text-sm leading-relaxed text-slate-700">
        Cuối 2024, mạng affiliate lớn gỡ Honey sau cáo buộc ghi đè hoa hồng (cookie stuffing / last-click
        hijack). Đây là rủi ro kiến trúc — không chỉ “bug PR”.
      </p>
      <ul className="mt-4 list-disc space-y-2 pl-5 text-sm text-slate-700">
        {SOURCES.map((s) => (
          <li key={s.href}>
            <a href={s.href} className="font-semibold text-sky-800 underline" rel="noreferrer" target="_blank">
              {s.label}
            </a>
          </li>
        ))}
      </ul>

      <h2 className="mt-10 text-xl font-black">Shopass khác bằng kiến trúc</h2>
      <ul className="mt-3 list-disc space-y-2 pl-5 text-sm text-slate-700">
        <li>Affiliate chỉ gắn khi bạn chủ động bấm — không inject ẩn.</li>
        <li>Extension công bố rõ dữ liệu đọc / không đọc (DISCLOSURE + /minh-bach).</li>
        <li>DNR allowlist + CI chặn host outbound mới (R31).</li>
        <li>Giá trị cốt lõi: lịch sử giá + sale ảo + cảnh báo đáy — không phải “coupon overlay”.</li>
      </ul>

      <p className="mt-6 text-sm text-slate-600">
        Chi tiết đầy đủ:{" "}
        <Link href="/minh-bach" className="font-bold text-sky-800 underline">
          trang Minh bạch
        </Link>
        .
      </p>

      <div className="mt-10 flex flex-wrap gap-3">
        <SignupCta className="rounded-xl bg-slate-950 px-5 py-3 text-sm font-extrabold text-white">
          Dùng Shopass thay Honey
        </SignupCta>
        <Link
          href="/bang-gia"
          className="rounded-xl border border-slate-200 px-5 py-3 text-sm font-bold text-slate-800"
        >
          Bảng giá
        </Link>
      </div>
    </article>
  );
}
