import type { Metadata } from "next";
import localFont from "next/font/local";
import { siteURL } from "@/lib/site-url";
import "./globals.css";

const geistSans = localFont({
  src: "./fonts/GeistVF.woff",
  variable: "--font-geist-sans",
  weight: "100 900",
});
const geistMono = localFont({
  src: "./fonts/GeistMonoVF.woff",
  variable: "--font-geist-mono",
  weight: "100 900",
});

export const metadata: Metadata = {
  metadataBase: new URL(siteURL),
  title: {
    default: "SănDeal - Theo dõi giá, tránh sale ảo",
    template: "%s | SănDeal",
  },
  description: "Theo dõi lịch sử giá và nhận cảnh báo deal tốt trên Shopee, TikTok Shop và Lazada.",
  applicationName: "SănDeal",
  alternates: { canonical: "/" },
  openGraph: {
    type: "website",
    locale: "vi_VN",
    siteName: "SănDeal",
    title: "SănDeal - Theo dõi giá, tránh sale ảo",
    description: "Mua đúng giá với lịch sử giá minh bạch và cảnh báo deal tốt.",
  },
  twitter: {
    card: "summary",
    title: "SănDeal - Theo dõi giá, tránh sale ảo",
    description: "Mua đúng giá với lịch sử giá minh bạch và cảnh báo deal tốt.",
  },
};

export const viewport = {
  width: "device-width",
  initialScale: 1,
  viewportFit: "cover",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="vi">
      <body className={`${geistSans.variable} ${geistMono.variable}`}>
        {children}
      </body>
    </html>
  );
}
