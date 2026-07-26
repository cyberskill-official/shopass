import type { Metadata } from "next";
import { PricingPage } from "@/components/pricing/pricing-page";
import { siteURL } from "@/lib/site-url";

export const metadata: Metadata = {
  title: "Bảng giá",
  description: "Free, Premium 29k, Plus 49k, Pro 79k — đăng ký chờ Premium Shopass.",
  alternates: { canonical: `${siteURL}/bang-gia` },
  openGraph: {
    title: "Bảng giá Shopass",
    description: "Free mạnh. Premium mở dự đoán đáy và wishlist lớn hơn.",
    url: `${siteURL}/bang-gia`,
    locale: "vi_VN",
  },
};

export default function BangGiaPage() {
  return <PricingPage />;
}
