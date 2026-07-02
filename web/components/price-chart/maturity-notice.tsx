import React from "react";
import type { Maturity } from "@/lib/chart/types";

export function MaturityNotice({ maturity }: { maturity: Maturity }) {
  if (maturity === "MATURE") return null;

  if (maturity === "NEW") {
    return (
      <div className="bg-gray-50 p-4 rounded-md text-sm text-gray-600 mb-4 border border-gray-200">
        Đang thu thập dữ liệu, chưa đủ để kết luận.
      </div>
    );
  }

  // WARMING
  return (
    <div className="bg-blue-50 p-3 rounded-md text-sm text-blue-700 mb-4 border border-blue-200">
      Đang tích lũy dữ liệu.
    </div>
  );
}
