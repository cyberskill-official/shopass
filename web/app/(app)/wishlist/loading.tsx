import { ListSkeleton } from "@/components/ui/list-skeleton";

export default function WishlistLoading() {
  return (
    <section className="mx-auto max-w-3xl space-y-4" aria-busy="true">
      <div className="h-10 w-64 animate-pulse rounded-xl bg-slate-100" />
      <ListSkeleton rows={5} />
    </section>
  );
}
