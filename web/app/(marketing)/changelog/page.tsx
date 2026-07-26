import type { Metadata } from "next";
import Link from "next/link";
import { CHANGELOG } from "@/lib/content/changelog";
import { siteURL } from "@/lib/site-url";

export const metadata: Metadata = {
  title: "Changelog Shopass",
  description: "Những gì vừa ship — bằng chứng sản phẩm còn sống.",
  alternates: { canonical: `${siteURL}/changelog` },
};

export default function ChangelogPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 py-12">
      <p className="text-sm font-bold text-sky-800">
        <Link href="/">Shopass</Link>
        {" · "}
        <Link href="/blog">Blog</Link>
      </p>
      <h1 className="mt-4 text-3xl font-black tracking-tight">Changelog</h1>
      <p className="mt-2 text-slate-600">Ghi chú phát hành ngắn. Chi tiết kỹ thuật nằm trong PR / LEDGER.</p>
      <ol className="mt-10 space-y-10">
        {CHANGELOG.map((entry) => (
          <li key={entry.version}>
            <p className="text-xs font-bold uppercase tracking-wide text-slate-500">
              {entry.date} · v{entry.version}
            </p>
            <h2 className="mt-1 text-xl font-black text-slate-900">{entry.title}</h2>
            <ul className="mt-3 list-disc space-y-1 pl-5 text-sm text-slate-700">
              {entry.items.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </li>
        ))}
      </ol>
    </div>
  );
}
