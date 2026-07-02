import React from "react";
import type { Verdict, Maturity } from "@/lib/chart/types";

const LABEL: Record<Verdict, string> = {
  SALE_AO: "Sale ảo",
  SALE_XIN: "Sale xịn",
  TAM_DUOC: "Tạm được",
  UNKNOWN: "Chưa đủ dữ liệu",
};

const COLOR: Record<Verdict, string> = {
  SALE_AO: "bg-red-100 text-red-800",
  SALE_XIN: "bg-green-100 text-green-800",
  TAM_DUOC: "bg-yellow-100 text-yellow-800",
  UNKNOWN: "bg-gray-100 text-gray-800",
};

export function VerdictBadge({ verdict, maturity }: { verdict: Verdict; maturity: Maturity }) {
  if (maturity === "NEW") return null;
  
  return (
    <span 
      data-verdict={verdict} 
      className={`inline-block px-2 py-1 text-xs font-semibold rounded-full ${COLOR[verdict]}`}
    >
      {LABEL[verdict]}
    </span>
  );
}
