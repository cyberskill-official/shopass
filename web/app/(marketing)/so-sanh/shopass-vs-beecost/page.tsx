import type { Metadata } from "next";
import Link from "next/link";
import { FeatureTable } from "@/components/compare/feature-table";
import { SignupCta } from "@/components/landing/landing-cta";
import { SHOPASS_VS_BEECOST } from "@/lib/compare/beecost";
import { siteURL } from "@/lib/site-url";

const title = "Shopass vs BeeCost — so sánh trung thực";
const description =
  "Bảng tính năng thực tế: TikTok Shop, lịch sử giá, dự đoán đáy, affiliate minh bạch. Không đả kích — chỉ sự thật tại ngày xuất bản.";

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: `${siteURL}/so-sanh/shopass-vs-beecost` },
  openGraph: {
    title,
    description,
    url: `${siteURL}/so-sanh/shopass-vs-beecost`,
    locale: "vi_VN",
    type: "article",
  },
};

const jsonLd = {
  "@context": "https://schema.org",
  "@type": "WebPage",
  name: title,
  description,
  url: `${siteURL}/so-sanh/shopass-vs-beecost`,
  inLanguage: "vi-VN",
  datePublished: "2026-07-26",
  dateModified: "2026-07-26",
};

export default function ShopassVsBeecostPage() {
  return (
    <article className="mx-auto max-w-4xl px-6 py-12 text-slate-900">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }} />
      <p className="text-sm font-bold text-sky-800">
        <Link href="/">Shopass</Link>
        {" · "}
        <Link href="/thay-the-honey">Thay thế Honey</Link>
        {" · "}
        <Link href="/minh-bach">Minh bạch</Link>
      </p>
      <h1 className="mt-4 text-3xl font-black tracking-tight sm:text-4xl">{title}</h1>
      <p className="mt-3 max-w-2xl text-slate-600">
        BeeCost mạnh về lịch sử giá đa sàn. Shopass bổ sung phát hiện sale ảo, dự đoán đáy (Premium), và
        kiến trúc affiliate chỉ khi bạn bấm. Bảng dưới cập nhật <strong>2026-07-26</strong> — nếu một
        ô sai, báo chúng tôi sửa.
      </p>

      <div className="mt-10">
        <FeatureTable rows={SHOPASS_VS_BEECOST} />
      </div>

      <p className="mt-6 text-xs text-slate-500">
        Nguồn nội bộ: PRD gating, TASK-DEAL fake-sale, /minh-bach, extension DISCLOSURE. Không dùng
        screenshot đối thủ có bản quyền trong trang này.
      </p>

      <div className="mt-10 flex flex-wrap gap-3">
        <SignupCta className="rounded-xl bg-slate-950 px-5 py-3 text-sm font-extrabold text-white">
          Thử Shopass miễn phí
        </SignupCta>
        <Link
          href="/kiem-tra-sale-ao"
          className="rounded-xl border border-slate-200 px-5 py-3 text-sm font-bold text-slate-800"
        >
          Kiểm tra sale ảo
        </Link>
      </div>
    </article>
  );
}
