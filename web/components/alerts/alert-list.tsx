import React from "react";
import type { AlertRule } from "@/lib/alerts/api";

const RULE_TYPE_LABELS: Record<string, string> = {
  price_below: "Giảm về giá",
  drop_pct: "Giảm %",
  real_sale: "Sale thật",
  bottom_predicted: "Sắp chạm đáy",
};

export function AlertList({
  alerts,
  onToggleActive,
  onDelete,
}: {
  alerts: AlertRule[];
  onToggleActive: (id: number, active: boolean) => Promise<void>;
  onDelete: (id: number) => Promise<void>;
}) {
  if (alerts.length === 0) {
    return (
      <div className="bg-gray-50 p-6 text-center rounded-lg border border-dashed border-gray-300">
        <p className="text-gray-500">Bạn chưa có cảnh báo nào. Hãy tạo một cảnh báo ở trên.</p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {alerts.map((alert) => (
        <div key={alert.id} className="bg-white p-4 rounded-lg border border-gray-200 flex justify-between items-center">
          <div>
            <div className="font-medium text-gray-800 mb-1">
              Sản phẩm #{alert.product_id}
            </div>
            <div className="text-sm text-gray-600 flex items-center space-x-2">
              <span className="bg-gray-100 px-2 py-0.5 rounded text-xs font-semibold">
                {RULE_TYPE_LABELS[alert.rule_type]}
              </span>
              {alert.threshold !== null && (
                <span>
                  {alert.rule_type === "price_below"
                    ? `${new Intl.NumberFormat("vi-VN").format(alert.threshold)} ₫`
                    : `${alert.threshold}%`}
                </span>
              )}
              <span className="text-gray-400">&bull;</span>
              <span className="uppercase text-xs">{alert.channels.join(", ")}</span>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <label className="flex items-center space-x-2 cursor-pointer">
              <span className="text-sm text-gray-600">{alert.active ? "Bật" : "Tắt"}</span>
              <div className="relative">
                <input
                  type="checkbox"
                  className="sr-only"
                  checked={alert.active}
                  onChange={async () => {
                    try {
                      await onToggleActive(alert.id, !alert.active);
                    } catch (error) {
                      window.alert(error instanceof Error && error.message ? error.message : "Đã xảy ra lỗi");
                    }
                  }}
                />
                <div className={`block w-10 h-6 rounded-full transition-colors ${alert.active ? "bg-blue-500" : "bg-gray-300"}`}></div>
                <div className={`absolute left-1 top-1 bg-white w-4 h-4 rounded-full transition-transform ${alert.active ? "transform translate-x-4" : ""}`}></div>
              </div>
            </label>

            <button
              onClick={async () => {
                if (confirm("Xóa cảnh báo này?")) {
                  try {
                    await onDelete(alert.id);
                  } catch (error) {
                    window.alert(error instanceof Error && error.message ? error.message : "Đã xảy ra lỗi");
                  }
                }
              }}
              className="text-red-500 hover:text-red-700 text-sm font-medium"
            >
              Xóa
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
