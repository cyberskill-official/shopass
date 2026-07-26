import type { Metadata } from "next";
import { LegalCompanyBlock, LegalCrossLinks, LegalDraftBanner } from "@/components/legal-doc";
import { siteURL } from "@/lib/site-url";

const title = "Chính sách bảo mật | Shopass";
const description =
  "Chính sách bảo mật Shopass (VN + EN) — dữ liệu thu thập, mục đích, retention dự kiến, DSAR.";

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: `${siteURL}/chinh-sach-bao-mat` },
  openGraph: {
    title,
    description,
    url: `${siteURL}/chinh-sach-bao-mat`,
    locale: "vi_VN",
    type: "article",
  },
};

export default function PrivacyPolicyPage() {
  return (
    <article lang="vi-VN" className="prose prose-slate mx-auto max-w-3xl px-6 py-12">
      <h1>Chính sách bảo mật</h1>
      <LegalDraftBanner />

      <p>
        Shopass (<a href="https://shopass.cyberskill.world">shopass.cyberskill.world</a>) giúp người
        mua Việt theo dõi lịch sử giá và nhận cảnh báo trên các sàn thương mại điện tử. Tài liệu này
        mô tả dữ liệu chúng tôi xử lý và quyền của bạn theo Luật Bảo vệ dữ liệu cá nhân (PDPL) Việt
        Nam — bản nháp chờ counsel.
      </p>

      <h2>1. Dữ liệu thu thập</h2>
      <ul>
        <li>Tài khoản: email/định danh đăng nhập, mã người dùng nội bộ.</li>
        <li>Sản phẩm theo dõi: URL/ID sản phẩm công khai, nền tảng (Shopee/TikTok/Lazada), lịch sử giá.</li>
        <li>
          Extension (khi bạn bật đồng bộ): platform, productId, giá hiển thị, số lượng giỏ, mã voucher
          công khai — xem{" "}
          <a href="https://github.com/cyberskill-official/shopass/blob/main/extension/DISCLOSURE.md">
            DISCLOSURE
          </a>
          .
        </li>
        <li>Nhật ký kỹ thuật tối thiểu: request id, mã lỗi, thời điểm (để vận hành và bảo mật).</li>
      </ul>

      <h2>2. Dữ liệu không bao giờ thu thập qua extension</h2>
      <ul>
        <li>Cookie phiên sàn, mật khẩu, token/header xác thực sàn.</li>
        <li>Email / SĐT / tên / địa chỉ lấy từ trang sàn (không gửi về server).</li>
      </ul>

      <h2>3. Mục đích</h2>
      <p>
        Cung cấp biểu đồ giá, cảnh báo giảm giá/đáy dự đoán, chống sale ảo ở mức sản phẩm, và vận hành
        dịch vụ (bảo mật, chống lạm dụng, hỗ trợ DSAR).
      </p>

      <h2>4. Lưu trữ (dự kiến — chờ chốt R19)</h2>
      <p>
        Lịch sử giá là xương sống sản phẩm; chính sách retention chi tiết sẽ được công bố sau quyết
        định sản phẩm. Dữ liệu tài khoản được giữ trong thời gian bạn dùng dịch vụ và trong thời hạn
        pháp lý bắt buộc sau khi xóa.
      </p>

      <h2>5. Quyền của bạn (DSAR)</h2>
      <p>
        Bạn có thể yêu cầu truy cập, chỉnh sửa, hoặc xóa dữ liệu cá nhân bằng email{" "}
        <a href="mailto:info@cyberskill.world">info@cyberskill.world</a>. Trang DSAR tự phục vụ sẽ
        được bổ sung (R32).
      </p>

      <h2>6. Tiếp thị liên kết</h2>
      <p>
        Shopass có thể hiển thị liên kết affiliate. Chúng tôi không gắn tham số affiliate khi bạn
        chưa chủ động bấm — không cookie-stuffing.
      </p>

      <h2 id="en">Privacy Policy (English)</h2>
      <p>
        Shopass processes account identifiers, publicly tracked product IDs/prices, and (with
        extension consent) cart/voucher display fields listed in our DISCLOSURE. We never collect
        marketplace session cookies, passwords, or auth tokens via the extension. Contact{" "}
        <a href="mailto:info@cyberskill.world">info@cyberskill.world</a> for DSAR requests. This
        English text is a summary; the Vietnamese sections above are primary. Draft pending counsel.
      </p>

      <LegalCrossLinks current="privacy" />
      <LegalCompanyBlock />
    </article>
  );
}
