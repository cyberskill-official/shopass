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
      setDeviceMsg("Đã đăng ký token push web.");
      setDeviceToken("");
    } catch (err) {
      setDeviceMsg(userMessage(err, "Không thể đăng ký token."));
    } finally {
      setDeviceBusy(false);
    }
  };

  return (
    <section className="mx-auto max-w-3xl space-y-8">
      <header>
        <h1 className="text-3xl font-black tracking-tight text-slate-900 sm:text-4xl">
          Cảnh báo giá
        </h1>
        <p className="mt-2 max-w-2xl text-base leading-relaxed text-slate-500">
          Tạo luật cảnh báo cho sản phẩm bạn theo dõi. Thông báo đẩy dùng token FCM web đã đăng ký.
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

      <form
        onSubmit={onRegisterDevice}
        className="space-y-3 rounded-lg border border-slate-200 bg-white p-6"
      >
        <h2 className="text-lg font-semibold text-slate-900">Đăng ký token push (web)</h2>
        <p className="text-sm text-slate-500">
          Dán FCM registration token (từ Firebase Messaging) để nhận push khi deal/alert bắn.
          Gateway sẽ publish <code className="text-xs">/v1/devices</code> cùng với alerts.
        </p>
        <input
          value={deviceToken}
          onChange={(e) => setDeviceToken(e.target.value)}
          placeholder="FCM token"
          className="w-full rounded border border-gray-300 px-3 py-2 font-mono text-sm"
        />
        <button
          type="submit"
          disabled={deviceBusy || !deviceToken.trim()}
          className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-extrabold text-white disabled:opacity-50"
        >
          {deviceBusy ? "Đang lưu…" : "Lưu token"}
        </button>
        {deviceMsg ? <p className="text-sm text-slate-600">{deviceMsg}</p> : null}
      </form>
    </section>
  );
}
