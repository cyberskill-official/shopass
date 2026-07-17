import { apiFetch } from "@/lib/api";
import { type RuleType, type Channel } from "./validate";

export interface AlertRule {
  id: number;
  product_id: number;
  rule_type: RuleType;
  threshold: number | null;
  channels: Channel[];
  active: boolean;
}

export interface AlertHistoryEntry {
  id: number;
  alert_rule_id: number;
  fired_at: string;
  payload: Record<string, unknown> | null;
  status: string;
}

export async function createAlert(data: {
  product_id: number;
  rule_type: RuleType;
  threshold: number | null;
  channels: Channel[];
}): Promise<AlertRule> {
  const res = await apiFetch("/v1/alerts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      product_id: data.product_id,
      rule_type: data.rule_type,
      threshold: data.threshold,
      channel: data.channels,
    }),
  });
  if (!res.ok) throw new Error("Không thể tạo cảnh báo");
  return (await res.json()) as AlertRule;
}

export async function listAlerts(): Promise<AlertRule[]> {
  const res = await apiFetch("/v1/alerts");
  if (!res.ok) throw new Error("Không thể tải danh sách cảnh báo");
  const rules = (await res.json()) as Array<
    Omit<AlertRule, "channels"> & { channel?: Channel[]; channels?: Channel[] }
  >;
  return rules.map(({ channel, channels, ...rule }) => ({
    ...rule,
    channels: channels ?? channel ?? [],
  }));
}

export async function toggleActive(id: number, active: boolean): Promise<void> {
  const res = await apiFetch(`/v1/alerts/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ active }),
  });
  if (!res.ok) throw new Error("Không thể thay đổi trạng thái cảnh báo");
}

export async function deleteAlert(id: number): Promise<void> {
  const res = await apiFetch(`/v1/alerts/${id}`, {
    method: "DELETE",
  });
  if (!res.ok) throw new Error("Không thể xóa cảnh báo");
}

export async function getAlertHistory(id: number): Promise<AlertHistoryEntry[]> {
  const res = await apiFetch(`/v1/alerts/${id}/history`);
  if (!res.ok) throw new Error("Không thể tải lịch sử cảnh báo");
  return (await res.json()) as AlertHistoryEntry[];
}
