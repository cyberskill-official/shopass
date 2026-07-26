export default function ChartLoading() {
  return (
    <div className="mx-auto max-w-5xl space-y-6" aria-busy="true" aria-label="Đang tải biểu đồ">
      <div className="h-8 w-40 animate-pulse rounded-lg bg-slate-100" />
      <div className="h-12 w-72 animate-pulse rounded-xl bg-slate-100" />
      <div className="h-[420px] animate-pulse rounded-[2rem] bg-slate-100" />
    </div>
  );
}
