import Link from "next/link";

export function LegalDraftBanner() {
  return (
    <p className="mb-8 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
      Bản nháp — chờ tư vấn pháp lý / Draft — pending legal counsel (R34 / R37).
    </p>
  );
}

export function LegalCompanyBlock() {
  return (
    <section className="mt-12 border-t border-slate-200 pt-8 text-sm leading-6 text-slate-600">
      <h2 className="text-base font-semibold text-slate-900">Đơn vị vận hành / Operator</h2>
      <p className="mt-2">
        CyberSkill Software Solutions Consultancy and Development JSC
        <br />
        1st Floor, 207A Nguyen Van Thu, Tan Dinh Ward, Ho Chi Minh City, Vietnam
        <br />
        Sản phẩm / Product:{" "}
        <a className="text-blue-700 underline" href="https://shopass.cyberskill.world">
          Shopass
        </a>
        <br />
        DSAR / liên hệ dữ liệu:{" "}
        <a className="text-blue-700 underline" href="mailto:info@cyberskill.world">
          info@cyberskill.world
        </a>
      </p>
    </section>
  );
}

export function LegalCrossLinks({ current }: { current: "privacy" | "terms" }) {
  return (
    <p className="mt-8 text-sm text-slate-600">
      {current === "privacy" ? (
        <>
          Xem thêm:{" "}
          <Link className="text-blue-700 underline" href="/dieu-khoan">
            Điều khoản sử dụng
          </Link>
        </>
      ) : (
        <>
          Xem thêm:{" "}
          <Link className="text-blue-700 underline" href="/chinh-sach-bao-mat">
            Chính sách bảo mật
          </Link>
        </>
      )}
    </p>
  );
}
