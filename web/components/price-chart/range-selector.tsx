import React from "react";
import { RANGE_ALLOWLIST, type Range } from "@/lib/chart/types";

export function RangeSelector({
  currentRange,
  onChange,
}: {
  currentRange: Range;
  onChange: (range: Range) => void;
}) {
  return (
    <div className="flex space-x-2 my-4 bg-gray-100 p-1 rounded-md inline-flex">
      {RANGE_ALLOWLIST.map((r) => (
        <button
          key={r}
          onClick={() => onChange(r)}
          className={`px-3 py-1 text-sm rounded-sm transition-colors ${
            currentRange === r
              ? "bg-white shadow-sm font-medium text-blue-600"
              : "text-gray-600 hover:text-gray-900"
          }`}
        >
          {r}
        </button>
      ))}
    </div>
  );
}
