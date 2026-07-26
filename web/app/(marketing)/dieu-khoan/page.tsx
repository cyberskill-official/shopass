import type { Metadata } from "next";
import { LegalCompanyBlock, LegalCrossLinks, LegalDraftBanner } from "@/components/legal-doc";
import { siteURL } from "@/lib/site-url";

const title = "Điều khoản sử dụng | Shopass";
const description =
  "Điều khoản sử dụng Shopass (VN + EN) — dịch vụ theo dõi giá, trách nhiệm, luật áp dụng.";

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: `${siteURL}/dieu-khoan` },
  openGraph: {
    title,
    description,
    url: `${siteURL}/dieu-khoan`,
    locale: "vi_VN",
    type: "article",
  },
};

export default function TermsPage() {
  return (
    <article lang="vi-VN" className="prose prose-slate mx-auto max-w-3xl px-6 py-12">
      <h1>Điều khoản sử dụng</h1>
      <LegalDraftBanner />

      <p>
        Bằng việc dùng Shopass (web hoặc extension), bạn đồng ý với các điều khoản dưới đây — bản
        nháp chờ tư vấn pháp lý.
      </p>

      <h2>1. Dịch vụ</h2>
      <p>
        Shopass cung cấp công cụ theo dõi giá, biểu đồ lịch sử, và cảnh báo. Đây không phải lời
        khuyên đầu tư hay bảo đảm mua được giá thấp nhất.
      </p>

      <h2>2. Tài khoản</h2>
      <p>
        Bạn chịu trách nhiệm bảo mật thông tin đăng nhập. Không dùng dịch vụ cho mục đích gian lận,
        phá hoại, hoặc thu thập dữ liệu trái phép.
      </p>

      <h2>3. Extension và sàn thương mại</h2>
      <p>
        Extension chỉ đọc dữ liệu hiển thị công khai trên trang bạn đang xem theo quyền đã cấp. Việc
        bạn tương tác với sàn (Shopee/TikTok/Lazada) chịu điều khoản của sàn đó.
      </p>

      <h2>4. Affiliate</h2>
      <p>
        Một số liên kết có thể mang hoa hồng. Shopass không gắn affiliate khi chưa có thao tác click
        rõ ràng của bạn.
      </p>

      <h2>5. Giới hạn trách nhiệm</h2>
      <p>
        Dịch vụ được cung cấp &quot;nguyên trạng&quot;. Trong phạm vi pháp luật cho phép, CyberSkill
        không chịu trách nhiệm cho thiệt hại gián tiếp phát sinh từ việc dùng hoặc không dùng được
        dịch vụ, hoặc từ sai lệch dữ liệu giá do thay đổi DOM/API của sàn.
      </p>

      <h2>6. Luật áp dụng</h2>
      <p>Điều khoản này được điều chỉnh bởi pháp luật Việt Nam.</p>

      <h2 id="en">Terms of Use (English)</h2>
      <p>
        Shopass provides price-tracking tools and alerts &quot;as is.&quot; You must not abuse the
        service. Affiliate parameters are only attached after an explicit user click. Governing law:
        Vietnam. English is a summary; Vietnamese sections are primary. Draft pending counsel.
      </p>

      <LegalCrossLinks current="terms" />
      <LegalCompanyBlock />
    </article>
  );
}
