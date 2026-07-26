import Link from "next/link";
import { LandingDemoChart } from "@/components/landing/demo-chart";
import { InstallCta, SignupCta } from "@/components/landing/landing-cta";
import { LANDING_FAQ, landingJsonLd } from "@/lib/landing/jsonld";

export function LandingPage() {
  const jsonLd = landingJsonLd();

  return (
    <main className="landing-root min-h-screen text-slate-900">
      {jsonLd.map((block, i) => (
        <script
          // eslint-disable-next-line react/no-danger
          key={i}
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(block) }}
        />
      ))}

      <nav className="sticky top-0 z-50 border-b border-slate-200/60 bg-[#f4f7fb]/80 backdrop-blur-xl">
        <div className="landing-container flex h-16 items-center justify-between lg:h-18">
          <Link href="/" className="text-lg font-black tracking-tight lg:text-xl">
            Shop<span className="text-sky-700">ass</span>
          </Link>
          <div className="flex items-center gap-4">
            <Link href="/login" className="hidden text-sm font-bold text-slate-600 hover:text-sky-800 sm:block">
              Đăng nhập
            </Link>
            <SignupCta className="rounded-xl bg-slate-950 px-4 py-2 text-sm font-bold text-white hover:bg-slate-800">
              Bắt đầu miễn phí
            </SignupCta>
          </div>
        </div>
      </nav>

      {/* Hero: brand-level name already in nav; headline + subline + CTAs + full-bleed chart */}
      <section className="relative overflow-hidden">
        <div className="landing-atmosphere" aria-hidden />
        <div className="landing-container relative px-6 pb-8 pt-14 sm:pt-16 lg:pt-20">
          <p className="text-sm font-extrabold uppercase tracking-[0.2em] text-sky-800">Shopass</p>
          <h1 className="mt-4 max-w-3xl text-4xl font-black leading-[1.08] tracking-tight text-slate-950 sm:text-5xl lg:text-6xl">
            Biết khi nào giá chạm đáy — trên Shopee, TikTok Shop, Lazada
          </h1>
          <p className="mt-5 max-w-xl text-base leading-relaxed text-slate-600 sm:text-lg">
            Phát hiện sale ảo bằng lịch sử giá thật. Cảnh báo khi gần đáy — không đoán mò.
          </p>
          <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:items-center">
            <SignupCta className="inline-flex items-center justify-center rounded-2xl bg-sky-700 px-6 py-4 text-sm font-extrabold text-white shadow-lg shadow-sky-200 hover:bg-sky-800">
              Theo dõi giá miễn phí
            </SignupCta>
            <InstallCta className="inline-flex items-center justify-center rounded-2xl border border-slate-300 bg-white/80 px-6 py-4 text-sm font-extrabold text-slate-800 hover:bg-white" />
          </div>
        </div>

        <div className="mt-10 w-full border-y border-slate-200/80 bg-white/70 px-2 py-4 sm:px-6 sm:py-6">
          <div className="landing-container">
            <p className="mb-2 text-xs font-bold uppercase tracking-wider text-slate-500">
              Ví dụ lịch sử giá (minh họa)
            </p>
            <LandingDemoChart />
          </div>
        </div>
      </section>

      <section id="how" className="landing-container px-6 py-16 lg:py-20">
        <h2 className="text-2xl font-black tracking-tight sm:text-3xl">Cách hoạt động</h2>
        <p className="mt-2 max-w-xl text-slate-600">Ba bước từ một đường link đến quyết định mua.</p>
        <ol className="mt-10 grid gap-10 md:grid-cols-3">
          {[
            ["Dán link", "Thêm sản phẩm Shopee / TikTok / Lazada vào bảng điều khiển."],
            ["Thu thập giá", "Shopass ghi nhận giá theo thời gian — lịch sử độc lập với 'giá gốc' trên sàn."],
            ["Nhận tín hiệu", "Biểu đồ + cảnh báo đáy / sale ảo khi đủ dữ liệu."],
          ].map(([t, d], i) => (
            <li key={t} className="relative pl-10">
              <span className="absolute left-0 top-0 text-2xl font-black text-sky-700/40">{i + 1}</span>
              <h3 className="text-lg font-black text-slate-900">{t}</h3>
              <p className="mt-2 text-sm leading-relaxed text-slate-600">{d}</p>
            </li>
          ))}
        </ol>
      </section>

      <section className="border-y border-slate-200/80 bg-white/50 px-6 py-12">
        <div className="landing-container">
          <h2 className="text-xl font-black">Niềm tin có thể kiểm chứng</h2>
          <ul className="mt-6 flex flex-col gap-4 text-sm font-semibold text-slate-700 sm:flex-row sm:flex-wrap sm:gap-8">
            <li>
              <Link className="text-sky-800 underline-offset-4 hover:underline" href="/minh-bach">
                Không cookie-stuffing
              </Link>
            </li>
            <li>
              <Link className="text-sky-800 underline-offset-4 hover:underline" href="/chinh-sach-bao-mat">
                PDPL / chính sách bảo mật
              </Link>
            </li>
            <li>
              <a
                className="text-sky-800 underline-offset-4 hover:underline"
                href="https://github.com/cyberskill-official/shopass"
                target="_blank"
                rel="noreferrer"
              >
                Mã nguồn trên GitHub
              </a>
              <span className="ml-1 font-normal text-slate-500">(extension MIT — R36 mirror sắp tới)</span>
            </li>
          </ul>
        </div>
      </section>

      <section id="faq" className="landing-container px-6 py-16 lg:py-20">
        <h2 className="text-2xl font-black tracking-tight sm:text-3xl">Câu hỏi thường gặp</h2>
        <div className="mt-8 divide-y divide-slate-200">
          {LANDING_FAQ.map((item) => (
            <details key={item.q} className="group py-5">
              <summary className="cursor-pointer list-none text-base font-bold text-slate-900 marker:content-none">
                {item.q}
              </summary>
              <p className="mt-3 max-w-2xl text-sm leading-relaxed text-slate-600">{item.a}</p>
            </details>
          ))}
        </div>
      </section>

      <section className="landing-container px-6 pb-20">
        <div className="rounded-[2rem] bg-slate-950 px-6 py-14 text-center text-white sm:px-12">
          <h2 className="text-2xl font-black sm:text-3xl">Sẵn sàng mua đúng lúc?</h2>
          <p className="mx-auto mt-4 max-w-lg text-sm text-slate-400">
            Tạo tài khoản miễn phí hoặc cài extension để bắt đầu theo dõi.
          </p>
          <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <SignupCta className="inline-flex rounded-2xl bg-sky-600 px-6 py-3 text-sm font-extrabold text-white hover:bg-sky-500">
              Tạo tài khoản
            </SignupCta>
            <InstallCta className="inline-flex rounded-2xl border border-slate-600 px-6 py-3 text-sm font-extrabold text-slate-100 hover:bg-slate-900" />
          </div>
        </div>
      </section>

      <footer className="border-t border-slate-200/70 px-6 py-10 text-sm text-slate-500">
        <div className="landing-container flex flex-col items-center justify-between gap-4 sm:flex-row">
          <p className="font-black text-slate-900">
            Shop<span className="text-sky-700">ass</span>
          </p>
          <p className="text-center sm:text-right">
            <Link href="/chinh-sach-bao-mat" className="hover:text-sky-800 hover:underline">
              Chính sách bảo mật
            </Link>
            {" · "}
            <Link href="/dieu-khoan" className="hover:text-sky-800 hover:underline">
              Điều khoản
            </Link>
            {" · "}
            <Link href="/minh-bach" className="hover:text-sky-800 hover:underline">
              Minh bạch
            </Link>
          </p>
        </div>
      </footer>
    </main>
  );
}
