"use client";

import { type FormEvent, useCallback, useEffect, useState } from "react";
import { AlertForm } from "@/components/alerts/alert-form";
import { AlertList } from "@/components/alerts/alert-list";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { RouteError } from "@/components/ui/route-error";
import {
  deleteAlert,
  listAlerts,
  registerDeviceToken,
  toggleActive,
  type AlertRule,
} from "@/lib/alerts/api";

function userMessage(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<AlertRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deviceToken, setDeviceToken] = useState("");
  const [deviceMsg, setDeviceMsg] = useState<string | null>(null);
  const [deviceBusy, setDeviceBusy] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setAlerts(await listAlerts());
    } catch (err) {
      setError(userMessage(err, "Không thể tải danh sách cảnh báo."));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const onToggleActive = async (id: number, active: boolean) => {
    await toggleActive(id, active);
    await refresh();
  };

  const onDelete = async (id: number) => {
    await deleteAlert(id);
    await refresh();
  };

  const onRegisterDevice = async (e: FormEvent) => {
    e.preventDefault();
    if (!deviceToken.trim()) return;
    setDeviceBusy(true);
    setDeviceMsg(null);
    try {
      await registerDeviceToken(deviceToken.trim(), "web");
      setDeviceMsg("Đã đăng ký thiết bị để nhận thông báo đẩy.");
      setDeviceToken("");
    } catch (err) {
      setDeviceMsg(userMessage(err, "Không thể đăng ký thiết bị."));
    } finally {
      setDeviceBusy(false);
    }
  };

  return (
    <section className="mx-auto max-w-3xl space-y-8">
      <header>
        <p className="mb-3 inline-flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider text-slate-400">
          <span className="h-1.5 w-1.5 rounded-full bg-sky-500" aria-hidden="true" />
          Closed beta
        </p>
        <h1 className="text-3xl font-black tracking-tight text-slate-900 sm:text-4xl">
          Cảnh báo giá
        </h1>
        <p className="mt-2 max-w-2xl text-base leading-relaxed text-slate-500">
          Tạo luật cảnh báo cho sản phẩm bạn theo dõi. Thông báo đẩy cần đăng ký thiết bị (phần bên
          dưới) trong closed beta.
        </p>
      </header>

      <AlertForm onCreated={() => void refresh()} />

      {loading ? (
        <ListSkeleton rows={4} />
      ) : error ? (
        <RouteError
          title="Không thể tải cảnh báo"
          message={error}
          onRetry={() => void refresh()}
        />
      ) : (
        <AlertList
          alerts={alerts}
          onToggleActive={onToggleActive}
          onDelete={onDelete}
        />
      )}

      <details className="group rounded-2xl border border-slate-200/80 bg-white p-6 shadow-sm shadow-slate-200/40">
        <summary className="cursor-pointer list-none text-lg font-black text-slate-900 marker:content-none">
          Đăng ký nhận thông báo đẩy
          <span className="mt-1 block text-sm font-medium text-slate-500">
            Tùy chọn nâng cao · token từ Firebase Messaging
          </span>
        </summary>
        <form onSubmit={onRegisterDevice} className="mt-5 space-y-3 border-t border-slate-100 pt-5">
          <p className="text-sm leading-relaxed text-slate-500">
            Dán mã đăng ký FCM của trình duyệt để nhận đẩy khi cảnh báo kích hoạt. Không bắt buộc
            để tạo luật — chỉ cần nếu bạn muốn thử push trên web.
          </p>
          <label htmlFor="fcm-token" className="block text-sm font-bold text-slate-700">
            Mã đăng ký thiết bị
          </label>
          <input
            id="fcm-token"
            value={deviceToken}
            onChange={(e) => setDeviceToken(e.target.value)}
            placeholder="Dán mã FCM…"
            autoComplete="off"
            className="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 font-mono text-sm text-slate-900 outline-none transition placeholder:font-sans placeholder:text-slate-400 focus:border-sky-400 focus:ring-4 focus:ring-sky-100"
          />
          <button
            type="submit"
            disabled={deviceBusy || !deviceToken.trim()}
            className="cursor-pointer rounded-xl bg-slate-950 px-4 py-2.5 text-sm font-extrabold text-white transition hover:bg-sky-700 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-sky-500/25 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {deviceBusy ? "Đang lưu…" : "Lưu thiết bị"}
          </button>
          {deviceMsg ? (
            <p className="text-sm font-medium text-slate-600" role="status">
              {deviceMsg}
            </p>
          ) : null}
        </form>
      </details>
    </section>
  );
}
