"use client";

export function RouteError({
  title,
  message,
  onRetry,
}: {
  title: string;
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div
      role="alert"
      className="flex flex-col items-center justify-center rounded-3xl border border-red-100 bg-red-50 px-6 py-16 text-center"
    >
      <h2 className="text-xl font-black text-red-900">{title}</h2>
      <p className="mt-2 max-w-md text-sm text-red-700">{message}</p>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="mt-6 rounded-xl border border-slate-200 bg-white px-5 py-2.5 text-sm font-bold text-slate-800"
        >
          Thử lại
        </button>
      )}
    </div>
  );
}
