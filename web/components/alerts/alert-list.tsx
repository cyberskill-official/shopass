import React from "react";
import Link from "next/link";
import type { AlertRule } from "@/lib/alerts/api";

const RULE_TYPE_LABELS: Record<string, string> = {
  price_below: "Giảm về giá",
  drop_pct: "Giảm %",
  real_sale: "Sale thật",
  bottom_predicted: "Sắp chạm đáy",
};

const CHANNEL_LABELS: Record<string, string> = {
  push: "Đẩy",
  email: "Email",
  sms: "SMS",
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
      <div className="flex flex-col items-center rounded-2xl border border-dashed border-slate-300/80 bg-gradient-to-b from-slate-50 to-white px-6 py-12 text-center">
        <div
          className="flex h-14 w-14 items-center justify-center rounded-2xl bg-white text-slate-400 shadow-sm ring-1 ring-slate-200/80"
          aria-hidden="true"
        >
          <svg className="h-7 w-7" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.75}
              d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
            />
          </svg>
        </div>
        <p className="mt-4 text-base font-black tracking-tight text-slate-900">
          Chưa có cảnh báo nào
        </p>
        <p className="mt-2 max-w-md text-sm leading-relaxed text-slate-500">
          Tạo luật ở form phía trên, hoặc bắt đầu từ hướng dẫn 2 phút để bật cảnh báo đầu tiên.
        </p>
        <Link
          href="/onboarding"
          className="mt-6 inline-flex cursor-pointer rounded-xl bg-slate-950 px-5 py-2.5 text-sm font-extrabold text-white shadow-sm transition hover:bg-sky-700 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-sky-500/25"
        >
          Bật cảnh báo đầu tiên (2 phút)
        </Link>
      </div>
    );
  }

  return (
    <ul className="space-y-3" aria-label="Danh sách cảnh báo">
      {alerts.map((alert) => (
        <li
          key={alert.id}
          className="flex flex-col gap-4 rounded-2xl border border-slate-200/80 bg-white p-4 shadow-sm shadow-slate-200/30 sm:flex-row sm:items-center sm:justify-between sm:p-5"
        >
          <div className="min-w-0">
            <div className="font-bold text-slate-900">
              <Link
                href={`/products/${alert.product_id}/chart`}
                className="cursor-pointer transition hover:text-sky-800 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-sky-500/20"
              >
                Sản phẩm #{alert.product_id}
              </Link>
            </div>
            <div className="mt-2 flex flex-wrap items-center gap-2 text-sm text-slate-600">
              <span className="rounded-lg bg-slate-100 px-2 py-0.5 text-xs font-bold text-slate-700">
                {RULE_TYPE_LABELS[alert.rule_type] ?? alert.rule_type}
              </span>
              {alert.threshold !== null && (
                <span className="font-medium">
                  {alert.rule_type === "price_below"
                    ? `${new Intl.NumberFormat("vi-VN").format(alert.threshold)} ₫`
                    : `${alert.threshold}%`}
                </span>
              )}
              <span className="text-slate-300" aria-hidden="true">
                ·
              </span>
              <span className="text-xs font-medium text-slate-500">
                {alert.channels.map((c) => CHANNEL_LABELS[c] ?? c).join(", ")}
              </span>
            </div>
          </div>

          <div className="flex shrink-0 items-center gap-4">
            <label className="flex cursor-pointer items-center gap-2">
              <span className="text-sm font-bold text-slate-600">
                {alert.active ? "Bật" : "Tắt"}
              </span>
              <span className="relative inline-flex">
                <input
                  type="checkbox"
                  className="peer sr-only"
                  checked={alert.active}
                  onChange={async () => {
                    try {
                      await onToggleActive(alert.id, !alert.active);
                    } catch (error) {
                      window.alert(
                        error instanceof Error && error.message
                          ? error.message
                          : "Đã xảy ra lỗi",
                      );
                    }
                  }}
                />
                <span
                  className={`block h-6 w-10 rounded-full transition-colors ${
                    alert.active ? "bg-sky-600" : "bg-slate-300"
                  }`}
                  aria-hidden="true"
                />
                <span
                  className={`absolute left-1 top-1 h-4 w-4 rounded-full bg-white shadow transition-transform ${
                    alert.active ? "translate-x-4" : ""
                  }`}
                  aria-hidden="true"
                />
              </span>
            </label>

            <button
              type="button"
              onClick={async () => {
                if (confirm("Xóa cảnh báo này?")) {
                  try {
                    await onDelete(alert.id);
                  } catch (error) {
                    window.alert(
                      error instanceof Error && error.message
                        ? error.message
                        : "Đã xảy ra lỗi",
                    );
                  }
                }
              }}
              className="cursor-pointer text-sm font-bold text-rose-600 transition hover:text-rose-800 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-rose-500/20"
            >
              Xóa
            </button>
          </div>
        </li>
      ))}
    </ul>
  );
}
