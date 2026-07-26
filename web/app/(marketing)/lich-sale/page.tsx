import type { Metadata } from "next";
import { SaleCalendarTool } from "@/components/tools/sale-calendar";
import { siteURL } from "@/lib/site-url";

export const metadata: Metadata = {
  title: "Lịch sale ngày đôi | Shopass",
  description: "Đếm ngược sale 1.1–12.12 và đăng ký nhắc email/Zalo.",
  alternates: { canonical: `${siteURL}/lich-sale` },
  openGraph: {
    title: "Lịch sale | Shopass",
    description: "Countdown ngày đôi + nhắc sale.",
    url: `${siteURL}/lich-sale`,
    locale: "vi_VN",
  },
};

export default function LichSalePage() {
  return <SaleCalendarTool />;
}
