import React from "react";
import type { Verdict, Maturity } from "@/lib/chart/types";

const LABEL: Record<Verdict, string> = {
  SALE_AO: "Sale ảo",
  SALE_XIN: "Sale xịn",
  TAM_DUOC: "Tạm được",
  UNKNOWN: "Chưa đủ dữ liệu",
};

const STYLES: Record<Verdict, string> = {
  SALE_AO: "bg-red-50 text-red-700 ring-red-200",
  SALE_XIN: "bg-emerald-50 text-emerald-700 ring-emerald-200",
  TAM_DUOC: "bg-amber-50 text-amber-700 ring-amber-200",
  UNKNOWN: "bg-slate-50 text-slate-600 ring-slate-200",
};

export function VerdictBadge({ verdict, maturity }: { verdict: Verdict; maturity: Maturity }) {
  if (maturity === "NEW") return null;

  return (
    <span
      data-verdict={verdict}
      className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-black uppercase tracking-wider ring-1 ring-inset ${STYLES[verdict]}`}
    >
      {verdict === "SALE_XIN" && (
        <span className="relative flex h-2 w-2">
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"></span>
          <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-500"></span>
        </span>
      )}
      {verdict === "SALE_AO" && (
        <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
      )}
      {LABEL[verdict]}
    </span>
  );
}
