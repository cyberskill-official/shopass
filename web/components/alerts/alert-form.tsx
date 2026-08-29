import React, { useState } from "react";
import { type RuleType, type Channel, CHANNELS, needsThreshold, validateAlert } from "@/lib/alerts/validate";
import { createAlert } from "@/lib/alerts/api";

const CHANNEL_LABELS: Record<Channel, string> = {
  push: "Đẩy (push)",
  email: "Email",
  sms: "SMS",
};

export function AlertForm({ onCreated }: { onCreated: () => void }) {
  const [productId, setProductId] = useState("");
  const [ruleType, setRuleType] = useState<RuleType>("price_below");
  const [threshold, setThreshold] = useState<string>("");
  const [channels, setChannels] = useState<Channel[]>(["push"]);
  const [loading, setLoading] = useState(false);

  const parsedThreshold = threshold.trim() === "" ? null : parseInt(threshold, 10);
  const error = validateAlert(ruleType, parsedThreshold, channels);

  const toggleChannel = (c: Channel) => {
    setChannels((prev) =>
      prev.includes(c) ? prev.filter((ch) => ch !== c) : [...prev, c]
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (error || !productId.trim()) return;

    setLoading(true);
    try {
      await createAlert({
        product_id: parseInt(productId, 10),
        rule_type: ruleType,
        threshold: parsedThreshold,
        channels: channels,
      });
      setProductId("");
      setThreshold("");
      setRuleType("price_below");
      setChannels(["push"]);
      onCreated();
    } catch (err) {
      alert(err instanceof Error && err.message ? err.message : "Đã xảy ra lỗi");
    } finally {
      setLoading(false);
    }
  };

  const fieldClass =
    "w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm font-medium text-slate-900 outline-none transition placeholder:text-slate-400 focus:border-sky-400 focus:ring-4 focus:ring-sky-100";

  return (
    <div className="rounded-2xl border border-slate-200/80 bg-white p-6 shadow-sm shadow-slate-200/40">
      <h2 className="text-lg font-black tracking-tight text-slate-900">Tạo luật cảnh báo</h2>
      <p className="mt-1 text-sm text-slate-500">
        Dùng mã sản phẩm từ Bảng điều khiển (số trên URL biểu đồ).
      </p>
      <form onSubmit={handleSubmit} className="mt-5 space-y-4">
        <div>
          <label htmlFor="alert-product-id" className="mb-1.5 block text-sm font-bold text-slate-700">
            Mã sản phẩm
          </label>
          <input
            id="alert-product-id"
            type="number"
            required
            value={productId}
            onChange={(e) => setProductId(e.target.value)}
            placeholder="Ví dụ 12345"
            className={fieldClass}
          />
        </div>

        <div>
          <label htmlFor="alert-rule-type" className="mb-1.5 block text-sm font-bold text-slate-700">
            Loại cảnh báo
          </label>
          <select
            id="alert-rule-type"
            value={ruleType}
            onChange={(e) => {
              setRuleType(e.target.value as RuleType);
              setThreshold("");
            }}
            className={fieldClass}
          >
            <option value="price_below">Báo khi về giá</option>
            <option value="drop_pct">Báo khi giảm %</option>
            <option value="real_sale">Báo khi sale thật</option>
            <option value="bottom_predicted">Báo khi sắp chạm đáy</option>
          </select>
        </div>

        {needsThreshold(ruleType) && (
          <div>
            <label htmlFor="alert-threshold" className="mb-1.5 block text-sm font-bold text-slate-700">
              Ngưỡng ({ruleType === "price_below" ? "VNĐ" : "%"})
            </label>
            <input
              id="alert-threshold"
              type="number"
              step={1}
              value={threshold}
              onChange={(e) => setThreshold(e.target.value)}
              className={fieldClass}
            />
          </div>
        )}

        <fieldset>
          <legend className="mb-2 text-sm font-bold text-slate-700">Kênh nhận (chọn nhiều)</legend>
          <div className="flex flex-wrap gap-3">
            {CHANNELS.map((c) => (
              <label
                key={c}
                className="inline-flex cursor-pointer items-center gap-2 rounded-xl border border-slate-200 bg-slate-50/80 px-3 py-2 text-sm font-medium text-slate-700 transition hover:bg-white has-[:checked]:border-sky-300 has-[:checked]:bg-sky-50 has-[:checked]:text-sky-900"
              >
                <input
                  type="checkbox"
                  checked={channels.includes(c)}
                  onChange={() => toggleChannel(c)}
                  className="rounded border-slate-300 text-sky-600 focus:ring-sky-500"
                />
                {CHANNEL_LABELS[c] ?? c}
              </label>
            ))}
          </div>
        </fieldset>

        {error && (
          <p className="text-sm font-medium text-rose-600" role="alert">
            {error}
          </p>
        )}

        <button
          type="submit"
          disabled={error != null || loading || !productId.trim()}
          className="w-full cursor-pointer rounded-xl bg-slate-950 py-3 text-sm font-extrabold text-white shadow-sm transition hover:bg-sky-700 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-sky-500/25 disabled:cursor-not-allowed disabled:bg-slate-300"
        >
          {loading ? "Đang tạo…" : "Tạo cảnh báo"}
        </button>
      </form>
    </div>
  );
}
