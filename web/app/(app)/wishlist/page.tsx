import Link from "next/link";

export default function WishlistPage() {
  return (
    <section className="mx-auto flex max-w-3xl flex-col items-center justify-center rounded-[2rem] border border-slate-200/60 bg-white p-10 text-center shadow-lg shadow-slate-200/50 sm:p-16">
      <div
        className="flex h-20 w-20 items-center justify-center rounded-3xl bg-slate-100 text-slate-500"
        aria-hidden="true"
      >
        <svg className="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.75}
            d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
          />
        </svg>
      </div>

      <p className="mt-8 inline-flex items-center gap-1.5 rounded-full bg-amber-50 px-3 py-1.5 text-xs font-bold uppercase tracking-wider text-amber-800 ring-1 ring-inset ring-amber-200/80">
        <span className="h-1.5 w-1.5 rounded-full bg-amber-500" aria-hidden="true" />
        Chưa mở · closed beta
      </p>

      <h1 className="mt-4 text-3xl font-black tracking-tight text-slate-900 sm:text-4xl">
        Danh sách quan tâm
      </h1>

      <p className="mx-auto mt-4 max-w-lg text-base leading-relaxed text-slate-500 sm:text-lg">
        Tính năng này chưa sẵn sàng — không phải lỗi 404. Trong closed beta, mọi sản phẩm bạn theo
        dõi nằm chung trên Bảng điều khiển. Wishlist (phân nhóm, nhiều danh sách) sẽ mở sau khi beta
        ổn định.
      </p>

      <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
        <Link
          href="/dashboard"
          className="inline-flex cursor-pointer items-center gap-2 rounded-xl bg-slate-900 px-6 py-3.5 text-sm font-extrabold text-white shadow-md transition hover:-translate-y-0.5 hover:bg-sky-700 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-sky-500/25"
        >
          Về Bảng điều khiển
        </Link>
        <Link
          href="/onboarding"
          className="inline-flex cursor-pointer items-center gap-2 rounded-xl border border-slate-200 bg-white px-6 py-3.5 text-sm font-bold text-slate-700 transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-blue-500/20"
        >
          Theo dõi sản phẩm đầu tiên
        </Link>
      </div>
    </section>
  );
}
