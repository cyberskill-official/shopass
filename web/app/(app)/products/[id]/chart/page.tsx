"use client";

import React, { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { fetchChart } from "@/lib/chart/fetch-chart";
import { PriceChart } from "@/components/price-chart/price-chart";
import { VerdictBadge } from "@/components/price-chart/verdict-badge";
import { MaturityNotice } from "@/components/price-chart/maturity-notice";
import { RangeSelector } from "@/components/price-chart/range-selector";
import type { ChartResponse, Range } from "@/lib/chart/types";

export default function ProductChartPage() {
  const params = useParams();
  const productId = Number(params.id);
  
  const [range, setRange] = useState<Range>("90d");
  const [data, setData] = useState<ChartResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    let ignore = false;
    
    async function loadData() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetchChart(productId, range);
        if (!ignore) {
          setData(res);
        }
      } catch (err: any) {
        if (!ignore) {
          setError(err.message || "Đã xảy ra lỗi khi tải biểu đồ");
        }
      } finally {
        if (!ignore) setLoading(false);
      }
    }
    
    if (productId && !isNaN(productId)) {
      loadData();
    }
    
    return () => { ignore = true; };
  }, [productId, range]);

  if (error) {
    return (
      <div className="p-6 bg-red-50 text-red-700 rounded-lg">
        <h2 className="font-bold mb-2">Lỗi</h2>
        <p>{error}</p>
      </div>
    );
  }

  return (
    <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100 max-w-4xl mx-auto">
      <div className="flex justify-between items-center mb-6 border-b pb-4">
        <h1 className="text-2xl font-bold text-gray-800">Lịch sử giá</h1>
        {data && (
          <VerdictBadge verdict={data.annotations.verdict} maturity={data.maturity} />
        )}
      </div>

      <RangeSelector currentRange={range} onChange={setRange} />
      
      {data && <MaturityNotice maturity={data.maturity} />}
      
      <div className="relative min-h-[320px]">
        {loading && !data && (
          <div className="absolute inset-0 flex items-center justify-center bg-white/80 z-10">
            <span className="text-gray-500">Đang tải biểu đồ...</span>
          </div>
        )}
        
        {data && <PriceChart data={data} />}
      </div>
    </div>
  );
}
