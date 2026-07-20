import Link from "next/link";
import { buildShopeeCaptureBookmarklet } from "@/lib/browser-capture";
import { siteURL } from "@/lib/site-url";

export default function CaptureGuidePage() {
  const bookmarklet = buildShopeeCaptureBookmarklet(siteURL);

  return (
    <main className="mx-auto max-w-3xl py-4 sm:py-8">
      <Link href="/dashboard" className="inline-flex items-center gap-1.5 text-sm font-bold text-slate-500 transition hover:text-slate-900">
        ← Về bảng điều khiển
      </Link>

      <section className="mt-5 rounded-[2rem] border border-slate-200 bg-white p-6 shadow-xl shadow-slate-200/50 sm:p-9">
        <p className="text-[10px] font-black uppercase tracking-[0.2em] text-blue-700">Công cụ browser-assisted</p>
        <h1 className="mt-3 text-2xl font-black tracking-tight text-slate-950 sm:text-3xl">Lấy giá Shopee trong một lần bấm</h1>
        <p className="mt-3 max-w-2xl text-sm leading-6 text-slate-600">
          Không cần proxy. Nút này chỉ chạy khi bạn chủ động bấm ở trang sản phẩm Shopee, rồi mở Shopass với link và giá đang hiển thị để bạn xác nhận.
        </p>

        <ol className="mt-7 space-y-4 text-sm leading-6 text-slate-700">
          <li className="flex gap-3"><span className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-blue-600 text-xs font-black text-white">1</span><span>Kéo nút bên dưới vào thanh dấu trang của trình duyệt. Không bấm nút khi đang ở Shopass.</span></li>
          <li className="flex gap-3"><span className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-blue-600 text-xs font-black text-white">2</span><span>Mở đúng sản phẩm và phân loại trên Shopee.</span></li>
          <li className="flex gap-3"><span className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-blue-600 text-xs font-black text-white">3</span><span>Bấm bookmark <strong>“Ghi giá Shopass”</strong>, kiểm tra lại giá rồi xác nhận ở Shopass.</span></li>
        </ol>

        <div className="mt-8 rounded-2xl border border-dashed border-blue-200 bg-blue-50/60 p-5 text-center">
          <a
            href={bookmarklet}
            draggable
            className="inline-flex cursor-grab items-center justify-center rounded-xl bg-blue-600 px-5 py-3 text-sm font-black text-white shadow-lg shadow-blue-200 transition hover:bg-blue-700 active:cursor-grabbing"
          >
            Ghi giá Shopass
          </a>
          <p className="mt-3 text-xs leading-5 text-slate-500">Nút không gửi cookie, mật khẩu, thông tin tài khoản hay lịch sử duyệt web. Nếu không đọc được giá, Shopass sẽ cho bạn nhập giá thủ công.</p>
        </div>
      </section>
    </main>
  );
}
