import Link from "next/link";

export default function AlertsPage() {
  return (
    <section className="mx-auto flex max-w-3xl flex-col items-center justify-center rounded-[2rem] border border-slate-200/60 bg-white p-10 text-center shadow-lg shadow-slate-200/50 sm:p-16">
      <div className="flex h-20 w-20 items-center justify-center rounded-3xl bg-amber-50 text-amber-500 shadow-inner">
        <svg className="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>
      </div>

      <p className="mt-8 inline-flex items-center gap-1.5 rounded-full bg-slate-100 px-3 py-1.5 text-xs font-bold uppercase tracking-wider text-slate-500">
        <span className="h-1.5 w-1.5 rounded-full bg-slate-400" />
        Sắp ra mắt
      </p>

      <h1 className="mt-4 text-3xl font-black tracking-tight text-slate-900 sm:text-4xl">Cảnh báo giá</h1>

      <p className="mx-auto mt-4 max-w-lg text-base leading-relaxed text-slate-500 sm:text-lg">
        Chúng tôi đang hoàn thiện hệ thống thông báo đa kênh. Tính năng này sẽ được mở khi đạt chuẩn độ tin cậy và tốc độ gửi tin. Hiện tại hệ thống đang tập trung thu thập dữ liệu minh bạch cho bạn.
      </p>

      <Link
        href="/dashboard"
        className="mt-8 inline-flex items-center gap-2 rounded-xl bg-slate-900 px-6 py-3.5 text-sm font-extrabold text-white shadow-md transition hover:-translate-y-0.5 hover:bg-blue-600"
      >
        Về Bảng điều khiển
      </Link>
    </section>
  );
}
