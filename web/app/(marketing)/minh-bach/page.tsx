import type { Metadata } from "next";
import Link from "next/link";
import { siteURL } from "@/lib/site-url";

const title = "Minh bạch | Shopass";
const description =
  "Shopass minh bạch dữ liệu extension, affiliate chỉ khi bạn bấm, và cách chúng tôi kiếm tiền.";

export const metadata: Metadata = {
  title,
  description,
  alternates: { canonical: `${siteURL}/minh-bach` },
  openGraph: {
    title,
    description,
    url: `${siteURL}/minh-bach`,
    locale: "vi_VN",
    type: "article",
  },
};

export default function TransparencyPage() {
  return (
    <article lang="vi-VN" className="prose prose-slate mx-auto max-w-3xl px-6 py-12">
      <h1>Minh bạch</h1>
      <p>
        Sau vụ Honey bị cắt affiliate vì ghi đè attribution ẩn, Shopass chọn kiến trúc đối lập: bạn
        thấy rõ dữ liệu nào rời máy, và affiliate chỉ gắn khi bạn chủ động bấm.
      </p>

      <h2>Extension đọc gì</h2>
      <ul>
        <li>Giá / ID sản phẩm công khai trên trang bạn đang xem</li>
        <li>Số lượng giỏ và mã voucher <em>hiển thị</em> (khi bạn bật đồng bộ)</li>
      </ul>
      <p>
        Chi tiết máy:{" "}
        <a href="https://github.com/cyberskill-official/shopass/blob/main/extension/DISCLOSURE.md">
          extension/DISCLOSURE.md
        </a>
        .
      </p>

      <h2>Không bao giờ đọc</h2>
      <ul>
        <li>Cookie phiên, mật khẩu, token sàn</li>
        <li>Email / SĐT / địa chỉ từ trang sàn</li>
        <li>Nội dung tab khác ngoài các host đã khai báo</li>
      </ul>

      <h2>Affiliate chỉ khi bạn bấm</h2>
      <p>
        Không cookie-stuffing. Guardrail tests trong repo chặn gắn tham số affiliate ngoài thao tác
        click tường minh (xem suite extension / affiliate khi R29 live).
      </p>

      <h2>Allowlist outbound</h2>
      <p>
        `host_permissions` + DNR allow rules đồng bộ với{" "}
        <code>extension/src/shared/allowlist.ts</code> (API:{" "}
        <code>api.shopass.cyberskill.world</code>, local{" "}
        <code>127.0.0.1:8080</code>). Thêm host mới mà không cập nhật allowlist sẽ làm fail CI (R31).
      </p>

      <h2>Cách Shopass kiếm tiền</h2>
      <ul>
        <li>Affiliate khi bạn click liên kết đủ điều kiện</li>
        <li>Gói Premium (cảnh báo / tính năng nâng cao)</li>
      </ul>

      <h2>Mã nguồn</h2>
      <p>
        Repo sản phẩm:{" "}
        <a href="https://github.com/cyberskill-official/shopass">cyberskill-official/shopass</a>.
        Extension MIT; public mirror tách repo chờ R36.
      </p>

      <h2 id="en">Transparency (English)</h2>
      <p>
        Shopass never exfiltrates marketplace session cookies or passwords. Affiliate parameters
        attach only after an explicit click. Outbound API hosts are CI-gated via the shared
        allowlist. Vietnamese sections above are primary.
      </p>

      <p className="not-prose mt-10 text-sm text-slate-600">
        <Link className="text-blue-700 underline" href="/chinh-sach-bao-mat">
          Chính sách bảo mật
        </Link>
        {" · "}
        <Link className="text-blue-700 underline" href="/dieu-khoan">
          Điều khoản
        </Link>
      </p>
    </article>
  );
}
