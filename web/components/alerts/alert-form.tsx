import React, { useState } from "react";
import { type RuleType, type Channel, CHANNELS, needsThreshold, validateAlert } from "@/lib/alerts/validate";
import { createAlert } from "@/lib/alerts/api";

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
    } catch (error) {
      alert(error instanceof Error && error.message ? error.message : "Đã xảy ra lỗi");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white p-6 rounded-lg border border-gray-200 mb-6">
      <h3 className="text-lg font-semibold mb-4">Tạo luật cảnh báo</h3>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">ID Sản phẩm</label>
          <input
            type="number"
            required
            value={productId}
            onChange={(e) => setProductId(e.target.value)}
            className="w-full border border-gray-300 rounded px-3 py-2"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Loại cảnh báo</label>
          <select
            value={ruleType}
            onChange={(e) => {
              setRuleType(e.target.value as RuleType);
              setThreshold(""); // reset threshold when changing rule type
            }}
            className="w-full border border-gray-300 rounded px-3 py-2 bg-white"
          >
            <option value="price_below">Báo khi về giá</option>
            <option value="drop_pct">Báo khi giảm %</option>
            <option value="real_sale">Báo khi sale thật</option>
            <option value="bottom_predicted">Báo khi sắp chạm đáy</option>
          </select>
        </div>

        {needsThreshold(ruleType) && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Ngưỡng ({ruleType === "price_below" ? "VND" : "%"})
            </label>
            <input
              type="number"
              step={1}
              value={threshold}
              onChange={(e) => setThreshold(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2"
            />
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Kênh nhận (chọn nhiều)</label>
          <div className="flex space-x-4">
            {CHANNELS.map((c) => (
              <label key={c} className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  checked={channels.includes(c)}
                  onChange={() => toggleChannel(c)}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <span className="text-sm text-gray-700 uppercase">{c}</span>
              </label>
            ))}
          </div>
        </div>

        {error && <p className="text-red-500 text-sm">{error}</p>}

        <button
          type="submit"
          disabled={error != null || loading || !productId.trim()}
          className="w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700 disabled:bg-gray-400"
        >
          {loading ? "Đang tạo..." : "Tạo cảnh báo"}
        </button>
      </form>
    </div>
  );
}
