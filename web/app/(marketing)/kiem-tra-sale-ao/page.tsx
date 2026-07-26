import type { Metadata } from "next";
import { FakeSaleChecker } from "@/components/tools/fake-sale-checker";
import { siteURL } from "@/lib/site-url";

export const metadata: Metadata = {
  title: "Kiểm tra sale ảo",
  description: "Dán link Shopee/Lazada/TikTok — xem sale thật hay ảo so với median 90 ngày.",
  alternates: { canonical: `${siteURL}/kiem-tra-sale-ao` },
  openGraph: {
    title: "Kiểm tra sale ảo | Shopass",
    description: "Verdict sale ảo miễn phí từ lịch sử giá Shopass.",
    url: `${siteURL}/kiem-tra-sale-ao`,
    locale: "vi_VN",
  },
};

export default function KiemTraSaleAoPage() {
  return <FakeSaleChecker />;
}
