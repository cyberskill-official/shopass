"use client";

import React, { useState, useEffect } from "react";
import { AlertForm } from "@/components/alerts/alert-form";
import { AlertList } from "@/components/alerts/alert-list";
import { listAlerts, toggleActive, deleteAlert, type AlertRule } from "@/lib/alerts/api";

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<AlertRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await listAlerts();
      setAlerts(data);
    } catch (e: any) {
      setError(e.message || "Đã xảy ra lỗi");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleToggle = async (id: number, active: boolean) => {
    await toggleActive(id, active);
    await loadData();
  };

  const handleDelete = async (id: number) => {
    await deleteAlert(id);
    await loadData();
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-2xl font-bold text-gray-800 mb-6">Cảnh báo giá</h1>
      
      <AlertForm onCreated={loadData} />

      <div className="mt-8">
        <h3 className="text-lg font-semibold mb-4">Danh sách cảnh báo</h3>
        {loading && alerts.length === 0 ? (
          <div className="text-gray-500">Đang tải...</div>
        ) : error ? (
          <div className="text-red-500">{error}</div>
        ) : (
          <AlertList 
            alerts={alerts} 
            onToggleActive={handleToggle} 
            onDelete={handleDelete} 
          />
        )}
      </div>
    </div>
  );
}
