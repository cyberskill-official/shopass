import Link from "next/link";

export default function WishlistPage() {
  return (
    <section className="mx-auto flex max-w-3xl flex-col items-center justify-center rounded-[2rem] border border-slate-200/60 bg-white p-10 text-center shadow-lg shadow-slate-200/50 sm:p-16">
      <div className="flex h-20 w-20 items-center justify-center rounded-3xl bg-blue-50 text-blue-500 shadow-inner">
        <svg className="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" /></svg>
      </div>

      <p className="mt-8 inline-flex items-center gap-1.5 rounded-full bg-slate-100 px-3 py-1.5 text-xs font-bold uppercase tracking-wider text-slate-500">
        <span className="h-1.5 w-1.5 rounded-full bg-slate-400" />
        Sắp ra mắt
      </p>

      <h1 className="mt-4 text-3xl font-black tracking-tight text-slate-900 sm:text-4xl">Danh sách quan tâm</h1>

      <p className="mx-auto mt-4 max-w-lg text-base leading-relaxed text-slate-500 sm:text-lg">
        Tính năng phân loại và quản lý danh sách sản phẩm yêu thích đang được phát triển. Trong thời gian dùng thử (Closed Beta), mọi sản phẩm bạn theo dõi sẽ được hiển thị chung trên Bảng điều khiển.
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
