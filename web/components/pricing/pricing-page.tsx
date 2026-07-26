"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { WaitlistModal } from "@/components/pricing/waitlist-modal";
import { trackEvent } from "@/lib/analytics";
import { isCheckoutLive } from "@/lib/billing/flags";

type TierKey = "free" | "premium_basic" | "premium_plus" | "premium_pro";

const TIERS: {
  key: TierKey;
  name: string;
  price: string;
  blurb: string;
  features: string[];
  cta: "signup" | "waitlist" | "checkout";
  waitlistTier?: "premium_basic" | "premium_plus" | "premium_pro";
}[] = [
  {
    key: "free",
    name: "Free",
    price: "0đ",
    blurb: "Theo dõi giá, sale ảo, biểu đồ — lõi miễn phí.",
    features: [
      "Theo dõi giá không giới hạn",
      "Phát hiện sale ảo",
      "Biểu đồ lịch sử giá",
      "Wishlist tối đa 20 mục",
      "Dự đoán đáy: chưa có",
    ],
    cta: "signup",
  },
  {
    key: "premium_basic",
    name: "Premium",
    price: "29.000đ/tháng",
    blurb: "Đủ dùng cho người săn deal hàng ngày.",
    features: [
      "Mọi thứ ở Free",
      "Wishlist tới 100 mục",
      "Dự đoán đáy (p_bottom)",
      "Cảnh báo nâng cao",
    ],
    cta: "waitlist",
    waitlistTier: "premium_basic",
  },
  {
    key: "premium_plus",
    name: "Premium Plus",
    price: "49.000đ/tháng",
    blurb: "Cho người theo dõi nhiều sản phẩm.",
    features: ["Wishlist tới 500 mục", "Dự đoán đáy", "Ưu tiên cảnh báo"],
    cta: "waitlist",
    waitlistTier: "premium_plus",
  },
  {
    key: "premium_pro",
    name: "Pro",
    price: "79.000đ/tháng",
    blurb: "Không giới hạn wishlist + quyền lợi cao nhất.",
    features: ["Wishlist không giới hạn", "Dự đoán đáy", "Hỗ trợ ưu tiên"],
    cta: "waitlist",
    waitlistTier: "premium_pro",
  },
];

const FAQ = [
  {
    q: "Khi nào thanh toán mở?",
    a: "Hiện đang thu danh sách chờ. Khi cổng MoMo/ZaloPay/VNPay sandbox (R28) sẵn sàng, CTA sẽ chuyển sang checkout thật.",
  },
  {
    q: "Free có bị khóa biểu đồ không?",
    a: "Không. Theo dõi giá, sale ảo và biểu đồ là tính năng lõi miễn phí (free-tier mạnh).",
  },
  {
    q: "Giá đã gồm VAT chưa?",
    a: "Giá niêm yết theo plan_catalog (BIGINT VND). Chi tiết hoá đơn sẽ hiện khi checkout live.",
  },
];

export function PricingPage() {
  const checkoutLive = isCheckoutLive();
  const [modalTier, setModalTier] = useState<"premium_basic" | "premium_plus" | "premium_pro" | null>(null);

  useEffect(() => {
    trackEvent("pricing-view");
  }, []);

  return (
    <main className="landing-root min-h-screen text-slate-900">
      <nav className="border-b border-slate-200/70 bg-white/70 px-6 py-4 backdrop-blur">
        <div className="landing-container flex items-center justify-between">
          <Link href="/" className="text-lg font-black">
            Shop<span className="text-sky-700">ass</span>
          </Link>
          <Link href="/login" className="text-sm font-bold text-slate-600 hover:text-sky-800">
            Đăng nhập
          </Link>
        </div>
      </nav>

      <section className="landing-container px-6 py-14">
        <h1 className="text-3xl font-black tracking-tight sm:text-4xl">Bảng giá Shopass</h1>
        <p className="mt-3 max-w-2xl text-slate-600">
          Free mạnh để săn deal. Premium mở dự đoán đáy và wishlist lớn hơn — đăng ký chờ nếu chưa
          thanh toán được.
        </p>

        <div className="mt-12 grid gap-6 lg:grid-cols-4">
          {TIERS.map((tier) => (
            <article
              key={tier.key}
              className="flex flex-col border border-slate-200 bg-white/80 p-6"
            >
              <h2 className="text-lg font-black">{tier.name}</h2>
              <p className="mt-2 text-2xl font-black text-sky-800">{tier.price}</p>
              <p className="mt-2 text-sm text-slate-600">{tier.blurb}</p>
              <ul className="mt-6 flex-1 space-y-2 text-sm text-slate-700">
                {tier.features.map((f) => (
                  <li key={f}>• {f}</li>
                ))}
              </ul>
              {tier.cta === "signup" && (
                <Link
                  href="/login?signup=1"
                  className="mt-8 inline-flex justify-center rounded-xl bg-slate-950 px-4 py-3 text-sm font-extrabold text-white"
                  onClick={() => trackEvent("tier-click", { tier: tier.key })}
                >
                  Bắt đầu miễn phí
                </Link>
              )}
              {tier.cta === "waitlist" && tier.waitlistTier ? (
                <button
                  type="button"
                  className="mt-8 rounded-xl bg-sky-700 px-4 py-3 text-sm font-extrabold text-white hover:bg-sky-800"
                  onClick={() => {
                    const waitlistTier = tier.waitlistTier;
                    if (!waitlistTier) return;
                    trackEvent("tier-click", { tier: tier.key });
                    if (checkoutLive) {
                      window.location.href = `/billing?tier=${waitlistTier}`;
                      return;
                    }
                    setModalTier(waitlistTier);
                  }}
                >
                  {checkoutLive ? "Thanh toán" : "Đăng ký chờ"}
                </button>
              ) : null}
            </article>
          ))}
        </div>
      </section>

      <section className="landing-container px-6 pb-16">
        <h2 className="text-xl font-black">Câu hỏi thanh toán</h2>
        <div className="mt-6 divide-y divide-slate-200">
          {FAQ.map((item) => (
            <details key={item.q} className="py-4">
              <summary className="cursor-pointer font-bold text-slate-900">{item.q}</summary>
              <p className="mt-2 text-sm text-slate-600">{item.a}</p>
            </details>
          ))}
        </div>
        <p className="mt-8 text-sm text-slate-500">
          <Link href="/chinh-sach-bao-mat" className="underline">
            Chính sách bảo mật
          </Link>
          {" · "}
          <Link href="/dieu-khoan" className="underline">
            Điều khoản
          </Link>
        </p>
      </section>

      {modalTier && (
        <WaitlistModal tier={modalTier} open onClose={() => setModalTier(null)} />
      )}
    </main>
  );
}
