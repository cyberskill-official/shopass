import React from "react";
import { RANGE_ALLOWLIST, type Range } from "@/lib/chart/types";

const LABELS: Record<Range, string> = {
  "7d": "7 ngày",
  "30d": "30 ngày",
  "90d": "90 ngày",
  "180d": "6 tháng",
  "1y": "1 năm",
};

export function RangeSelector({
  currentRange,
  onChange,
}: {
  currentRange: Range;
  onChange: (range: Range) => void;
}) {
  return (
    <div className="inline-flex items-center rounded-xl bg-slate-100 p-1 shadow-inner">
      {RANGE_ALLOWLIST.map((r) => {
        const isActive = currentRange === r;
        return (
          <button
            key={r}
            onClick={() => onChange(r)}
            className={`relative rounded-lg px-3.5 py-1.5 text-xs font-bold transition-all duration-200 sm:px-4 sm:text-sm ${
              isActive
                ? "bg-white text-slate-900 shadow-sm ring-1 ring-slate-200/50"
                : "text-slate-500 hover:text-slate-700"
            }`}
          >
            {LABELS[r] || r}
          </button>
        );
      })}
    </div>
  );
}
