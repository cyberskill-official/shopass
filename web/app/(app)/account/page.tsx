"use client";

import Link from "next/link";
import { useState } from "react";
import { apiFetch } from "@/lib/api";
import { logout } from "@/lib/auth";

export default function AccountPage() {
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [graceUntil, setGraceUntil] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const onDelete = async () => {
    if (confirm !== "XÓA") {
      setError('Gõ chính xác XÓA để xác nhận.');
      return;
    }
    setBusy(true);
    setError("");
    try {
      const res = await apiFetch("/v1/account", { method: "DELETE" });
      const body = await res.json().catch(() => null);
      if (!res.ok) {
        setError(body?.error || "Không xóa được tài khoản");
        return;
      }
      setGraceUntil(typeof body?.grace_until === "string" ? body.grace_until : null);
      await logout();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Không xóa được tài khoản");
    } finally {
      setBusy(false);
    }
  };

  if (graceUntil) {
    return (
      <div className="mx-auto max-w-lg space-y-4 rounded-3xl border border-slate-200 bg-white p-8">
        <h1 className="text-2xl font-black text-slate-950">Tài khoản đã xóa</h1>
        <p className="text-sm text-slate-600">
          Dữ liệu định danh đã được ẩn danh hóa (DSAR PDPL). Bạn không thể đăng nhập lại. Cửa sổ ân hạn
          (purge cứng) dự kiến tới{" "}
          <time dateTime={graceUntil}>{new Date(graceUntil).toLocaleString("vi-VN")}</time>.
        </p>
        <Link href="/" className="inline-block text-sm font-bold text-sky-800 underline">
          Về trang chủ
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg space-y-8">
      <header>
        <h1 className="text-3xl font-black tracking-tight text-slate-950">Tài khoản</h1>
        <p className="mt-2 text-sm text-slate-600">
          Closed beta: quản lý tối thiểu — đặt lại mật khẩu qua email (cần SMTP/R23) và xóa tài khoản
          theo quyền DSAR.
        </p>
      </header>

      <section className="rounded-3xl border border-slate-200 bg-white p-6">
        <h2 className="text-lg font-black text-slate-900">Mật khẩu</h2>
        <p className="mt-2 text-sm text-slate-600">
          Dùng trang quên mật khẩu khi bạn không nhớ mật khẩu hiện tại.
        </p>
        <Link
          href="/forgot-password"
          className="mt-4 inline-block rounded-xl border border-slate-200 px-4 py-2 text-sm font-bold text-slate-800 hover:bg-slate-50"
        >
          Quên mật khẩu
        </Link>
      </section>

      <section className="rounded-3xl border border-rose-200 bg-rose-50/40 p-6">
        <h2 className="text-lg font-black text-rose-950">Xóa tài khoản (DSAR)</h2>
        <p className="mt-2 text-sm text-rose-900/90">
          Thao tác không hoàn tác từ giao diện: email/phone bị ẩn danh, phiên bị thu hồi, status =
          deleted. Có cửa sổ ân hạn ~14 ngày trước purge cứng (ops).
        </p>
        <label className="mt-4 block text-xs font-extrabold uppercase tracking-wide text-rose-800">
          Gõ XÓA để xác nhận
          <input
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            className="mt-2 w-full rounded-xl border border-rose-200 bg-white px-3 py-2 text-sm"
            autoComplete="off"
          />
        </label>
        {error && (
          <p role="alert" className="mt-3 text-sm font-medium text-rose-800">
            {error}
          </p>
        )}
        <button
          type="button"
          disabled={busy}
          onClick={() => void onDelete()}
          className="mt-4 rounded-xl bg-rose-700 px-4 py-3 text-sm font-extrabold text-white hover:bg-rose-800 disabled:opacity-60"
        >
          {busy ? "Đang xóa…" : "Xóa tài khoản vĩnh viễn"}
        </button>
      </section>
    </div>
  );
}
