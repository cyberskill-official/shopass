import Link from "next/link";
import React from "react";

export default function Home() {
  return (
    <main className="mesh-bg min-h-screen text-slate-900">
      {/* Sticky Header */}
      <nav className="sticky top-0 z-50 border-b border-slate-200/50 bg-white/60 backdrop-blur-xl">
        <div className="landing-container flex h-16 items-center justify-between lg:h-20">
          <Link href="/" className="flex items-center gap-3 text-lg font-black tracking-tight lg:text-xl">
            <span className="grid h-9 w-9 place-items-center rounded-[14px] bg-gradient-to-br from-blue-600 to-violet-600 text-white shadow-lg shadow-blue-200 lg:h-10 lg:w-10 lg:rounded-2xl">S</span>
            <span>Shop<span className="text-blue-600">ass</span></span>
          </Link>
          <div className="flex items-center gap-4">
            <Link href="/login" className="hidden text-sm font-bold text-slate-600 transition hover:text-blue-600 sm:block">Đăng nhập</Link>
            <Link href="/login?signup=1" className="rounded-xl bg-slate-950 px-4 py-2 text-sm font-bold text-white shadow-lg shadow-slate-200 transition hover:-translate-y-0.5 hover:bg-slate-800 lg:px-5 lg:py-2.5">
              Bắt đầu <span className="hidden sm:inline">miễn phí</span>
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="landing-container landing-hero">
        <div className="landing-copy order-2 lg:order-1">
          <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-blue-100 bg-white/80 px-3 py-1.5 text-xs font-extrabold uppercase tracking-[.18em] text-blue-700 shadow-sm backdrop-blur-md">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-500"></span>
            </span>
            Price intelligence cho người mua Việt
          </div>

          <h1 className="font-black leading-[1.05] tracking-tight text-slate-950">
            Mua đúng lúc.<br />
            <span className="bg-gradient-to-r from-blue-600 via-indigo-600 to-violet-600 bg-clip-text text-transparent">Không mua hớ.</span>
          </h1>

          <p className="mt-6 max-w-xl text-base leading-relaxed text-slate-600 sm:text-lg sm:leading-8">
            Shopass biến lịch sử giá thành một quyết định đơn giản: hôm nay là lúc nên mua ngay, hay bạn nên chờ thêm?
          </p>

          <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:items-center lg:mt-10 lg:gap-4">
            <Link href="/login?signup=1" className="flex items-center justify-center gap-2 rounded-2xl bg-blue-600 px-6 py-4 text-sm font-extrabold text-white shadow-xl shadow-blue-200 transition hover:-translate-y-0.5 hover:bg-blue-700">
              Bắt đầu miễn phí <span>→</span>
            </Link>
            <a href="#how" className="flex items-center justify-center rounded-2xl border border-slate-200 bg-white/80 px-6 py-4 text-sm font-extrabold text-slate-700 transition hover:bg-white hover:text-slate-900">
              Xem cách hoạt động
            </a>
          </div>

          <div className="mt-10 grid grid-cols-3 gap-4 border-t border-slate-200/50 pt-8 sm:gap-8 lg:mt-12 lg:pt-10">
            <div className="text-sm text-slate-500">
              <strong className="block text-xl font-black text-slate-900 sm:text-2xl">24/7</strong>
              Thu thập giá
            </div>
            <div className="text-sm text-slate-500">
              <strong className="block text-xl font-black text-slate-900 sm:text-2xl">100%</strong>
              Minh bạch
            </div>
            <div className="text-sm text-slate-500">
              <strong className="block text-xl font-black text-slate-900 sm:text-2xl">1 link</strong>
              Để bắt đầu
            </div>
          </div>
        </div>

        {/* Mockup Showcase */}
        <div className="landing-showcase relative order-1 mx-auto w-full max-w-md lg:order-2 lg:max-w-none">
          <div className="absolute -inset-8 rounded-[3rem] bg-blue-400/20 blur-3xl lg:-inset-12" />
          <div className="relative w-full overflow-hidden rounded-[2rem] border border-white/80 bg-slate-950 p-3 shadow-2xl shadow-indigo-200/50 sm:p-4">
            <div className="min-w-0 rounded-[1.4rem] bg-white p-4 sm:p-6">
              <div className="flex min-w-0 items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <p className="text-xs font-bold text-slate-400">SẢN PHẨM ĐANG THEO DÕI</p>
                  <p className="mt-1 truncate text-base font-black sm:text-lg">Sony WH-1000XM5</p>
                </div>
                <span className="shrink-0 rounded-full bg-emerald-50 px-3 py-1.5 text-xs font-bold text-emerald-700 ring-1 ring-emerald-200/50">Sale xịn</span>
              </div>

              <div className="mt-6 flex min-w-0 flex-wrap items-end justify-between gap-4 sm:mt-8">
                <div>
                  <p className="text-xs font-medium text-slate-400">Giá hiện tại</p>
                  <p className="text-3xl font-black tracking-tight sm:text-4xl">6.490.000₫</p>
                  <p className="mt-1.5 flex items-center gap-1 text-sm font-bold text-emerald-600">
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M19 14l-7 7m0 0l-7-7m7 7V3" /></svg>
                    18,4% so với đỉnh 90 ngày
                  </p>
                </div>
                <div className="rounded-xl border border-blue-100 bg-blue-50/50 px-3 py-2 text-right">
                  <p className="text-xs font-bold text-blue-700">Đáy 90 ngày</p>
                  <p className="text-base font-black text-blue-900 sm:text-lg">6.190.000₫</p>
                </div>
              </div>

              <div className="mt-8 flex h-32 items-end gap-1.5 rounded-2xl bg-gradient-to-b from-slate-50 to-white px-2 pb-4 pt-8 sm:h-40 sm:gap-2 sm:px-4">
                {[45, 55, 48, 62, 70, 75, 60, 48, 55, 30, 40, 25, 35, 18].map((h, i) => (
                  <div
                    key={i}
                    className={`min-w-0 flex-1 rounded-t-sm transition-all duration-500 sm:rounded-t-md ${i > 10 ? "bg-blue-600" : "bg-blue-200"}`}
                    style={{ height: `${h}%` }}
                  />
                ))}
              </div>
              <p className="mt-4 text-center text-xs font-semibold text-slate-400">Lịch sử giá minh họa · 90 ngày gần nhất</p>
            </div>
          </div>
        </div>
      </section>

      {/* How it works */}
      <section id="how" className="landing-container py-16 lg:py-24">
        <p className="text-center text-xs font-black uppercase tracking-[.2em] text-blue-600">
          Đơn giản để dùng · Mạnh mẽ để quyết định
        </p>
        <h2 className="mx-auto mt-4 max-w-2xl text-center text-3xl font-black tracking-tight sm:text-4xl lg:text-5xl">
          Từ một đường link đến quyết định thông minh
        </h2>

        <div className="mt-16 grid gap-6 md:grid-cols-3 lg:gap-8">
          {[
            {
              n: "01",
              t: "Dán link sản phẩm",
              d: "Copy link trực tiếp của sản phẩm trên Shopee và dán vào Dashboard của bạn. Chỉ mất vài giây để bắt đầu.",
              icon: <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" /></svg>
            },
            {
              n: "02",
              t: "Hệ thống thu thập",
              d: "Shopass tự động ghi nhận giá mỗi ngày và xây dựng một lịch sử giá độc lập, trung thực, hoàn toàn minh bạch.",
              icon: <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" /></svg>
            },
            {
              n: "03",
              t: "Quyết định thông minh",
              d: "Biểu đồ trực quan và tín hiệu AI cho bạn biết giá hiện tại đang tốt đến mức nào — không cần phải đoán mò.",
              icon: <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>
            }
          ].map(({ n, t, d, icon }) => (
            <div key={n} className="group relative flex flex-col rounded-3xl border border-white/80 bg-white/60 p-8 shadow-sm backdrop-blur-sm transition-all hover:bg-white hover:shadow-xl hover:shadow-blue-100/50">
              <div className="mb-6 flex h-14 w-14 items-center justify-center rounded-2xl bg-blue-50 text-blue-600 shadow-inner group-hover:bg-blue-600 group-hover:text-white transition-colors">
                {icon}
              </div>
              <span className="absolute right-8 top-8 text-4xl font-black text-slate-100 transition-colors group-hover:text-blue-50">{n}</span>
              <h3 className="text-xl font-black text-slate-900">{t}</h3>
              <p className="mt-4 flex-1 text-sm leading-relaxed text-slate-600">{d}</p>
            </div>
          ))}
        </div>
      </section>

      {/* CTA Section */}
      <section className="landing-container mb-24">
        <div className="relative overflow-hidden rounded-[2.5rem] bg-slate-950 px-6 py-16 text-center shadow-2xl shadow-slate-300 sm:px-12 sm:py-20 lg:py-24">
          <div className="absolute -left-20 -top-20 h-64 w-64 rounded-full bg-blue-600/30 blur-3xl" />
          <div className="absolute -bottom-20 -right-20 h-64 w-64 rounded-full bg-violet-600/30 blur-3xl" />
          <div className="relative mx-auto max-w-2xl">
            <h2 className="text-3xl font-black tracking-tight text-white sm:text-4xl lg:text-5xl">Ngừng mua hớ ngay hôm nay.</h2>
            <p className="mt-6 text-base leading-relaxed text-slate-400 sm:text-lg">Tạo tài khoản miễn phí để bắt đầu theo dõi giá những món đồ bạn yêu thích. Bạn chỉ việc thêm link, Shopass sẽ lo phần còn lại.</p>
            <Link href="/login?signup=1" className="mx-auto mt-10 inline-flex items-center justify-center gap-2 rounded-2xl bg-blue-600 px-8 py-4 text-base font-extrabold text-white shadow-xl shadow-blue-900/50 transition hover:-translate-y-0.5 hover:bg-blue-500">
              Tạo tài khoản miễn phí <span>→</span>
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-slate-200/60 bg-white/40 px-6 py-10">
        <div className="landing-container flex flex-col items-center justify-between gap-6 sm:flex-row">
          <div className="flex items-center gap-2 font-black text-slate-900">
            <span className="grid h-7 w-7 place-items-center rounded-lg bg-slate-900 text-xs text-white">S</span>
            Shopass
          </div>
          <p className="text-center text-sm font-medium text-slate-500 sm:text-right">
            Sản phẩm đang trong giai đoạn Closed Beta.<br className="sm:hidden" /> Dữ liệu giá minh bạch cho người dùng Việt Nam.
          </p>
        </div>
      </footer>
    </main>
  );
}
