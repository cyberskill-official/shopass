import { ListSkeleton } from "@/components/ui/list-skeleton";

export default function AlertsLoading() {
  return (
    <section className="mx-auto max-w-3xl space-y-4" aria-busy="true">
      <div className="h-10 w-48 animate-pulse rounded-xl bg-slate-100" />
      <div className="h-40 animate-pulse rounded-xl bg-slate-100" />
      <ListSkeleton rows={4} />
    </section>
  );
}
