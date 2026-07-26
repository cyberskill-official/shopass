"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { trackEvent } from "@/lib/analytics";
import { getReferralMe, type ReferralMe } from "@/lib/referral/api";

export function ReferralCard() {
  const [data, setData] = useState<ReferralMe | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [origin, setOrigin] = useState("");

  const load = useCallback(async () => {
    try {
      setData(await getReferralMe());
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Lỗi");
    }
  }, []);

  useEffect(() => {
    setOrigin(window.location.origin);
    void load();
  }, [load]);

  if (error) {
    return (
      <section className="rounded-3xl border border-dashed border-slate-200 bg-slate-50 p-6 text-sm text-slate-600">
        Chưa tải được mã giới thiệu. {error}
      </section>
    );
  }

  if (!data) {
    return (
      <section className="animate-pulse rounded-3xl border border-slate-200 bg-white p-6">
        <div className="h-4 w-40 rounded bg-slate-100" />
      </section>
    );
  }

  const shareURL = `${origin}/login?signup=1&next=/onboarding&ref=${encodeURIComponent(data.code)}`;

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(shareURL);
      setCopied(true);
      trackEvent("share-click", { channel: "copy", code: data.code });
      window.setTimeout(() => setCopied(false), 2000);
    } catch {
      setCopied(false);
    }
  };

  const zaloHref = `https://zalo.me/?text=${encodeURIComponent(`Dùng Shopass săn deal thật — đăng ký với mã của mình: ${shareURL}`)}`;

  return (
    <section className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm sm:p-8">
      <h2 className="text-xl font-black text-slate-900">Mời bạn — cả hai nhận Premium</h2>
      <p className="mt-2 text-sm text-slate-600">{data.reward_note}</p>
      <p className="mt-4 font-mono text-2xl font-black tracking-widest text-sky-900">{data.code}</p>
      <p className="mt-1 text-xs font-bold text-slate-500">Đã giới thiệu thành công: {data.uses}</p>
      <div className="mt-5 flex flex-wrap gap-2">
        <button
          type="button"
          onClick={() => void copy()}
          className="rounded-xl bg-slate-950 px-4 py-2.5 text-sm font-extrabold text-white"
        >
          {copied ? "Đã copy link" : "Copy link mời"}
        </button>
        <a
          href={zaloHref}
          target="_blank"
          rel="noreferrer"
          onClick={() => trackEvent("share-click", { channel: "zalo", code: data.code })}
          className="rounded-xl border border-slate-200 px-4 py-2.5 text-sm font-bold text-slate-800"
        >
          Chia sẻ Zalo
        </a>
        <Link href="/dieu-khoan" className="rounded-xl px-4 py-2.5 text-sm font-bold text-sky-800 underline">
          Điều khoản
        </Link>
      </div>
    </section>
  );
}
