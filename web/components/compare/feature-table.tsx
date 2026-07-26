import { CELL_LABEL, type CompareRow } from "@/lib/compare/beecost";

export function FeatureTable({ rows }: { rows: CompareRow[] }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[36rem] border-collapse text-left text-sm">
        <thead>
          <tr className="border-b border-slate-200 text-xs font-bold uppercase tracking-wide text-slate-500">
            <th className="py-3 pr-4">Tính năng</th>
            <th className="py-3 pr-4">Shopass</th>
            <th className="py-3 pr-4">BeeCost</th>
            <th className="py-3">Ghi chú</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.feature} className="border-b border-slate-100 align-top">
              <td className="py-3 pr-4 font-semibold text-slate-900">{row.feature}</td>
              <td className="py-3 pr-4 font-bold text-sky-900">{CELL_LABEL[row.shopass]}</td>
              <td className="py-3 pr-4 font-bold text-slate-700">{CELL_LABEL[row.beecost]}</td>
              <td className="py-3 text-slate-500">{row.note ?? "—"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
