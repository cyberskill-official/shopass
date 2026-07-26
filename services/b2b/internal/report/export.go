package report

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
)

// ExportJSON renders aggregate-only report rows (no identity keys).
func ExportJSON(r Report) ([]byte, error) {
	type row struct {
		CategoryID     int64   `json:"category_id"`
		PlatformID     int16   `json:"platform_id"`
		Day            string  `json:"day"`
		MedianP        int64   `json:"median_p"`
		P25P           int64   `json:"p25_p"`
		P75P           int64   `json:"p75_p"`
		AvgDiscountPct float64 `json:"avg_discount_pct"`
		SKUCount       int32   `json:"sku_count"`
	}
	out := make([]row, 0, len(r.Cells))
	for _, c := range r.Cells {
		out = append(out, row{
			CategoryID: c.CategoryID, PlatformID: c.PlatformID, Day: c.Day,
			MedianP: c.MedianP, P25P: c.P25P, P75P: c.P75P,
			AvgDiscountPct: c.AvgDiscountPct, SKUCount: c.SKUCount,
		})
	}
	return json.Marshal(out)
}

// ExportCSV renders aggregate-only CSV (DEC-B2B-14).
func ExportCSV(r Report) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{
		"category_id", "platform_id", "day",
		"median_p", "p25_p", "p75_p", "avg_discount_pct", "sku_count",
	}); err != nil {
		return "", err
	}
	for _, c := range r.Cells {
		if err := w.Write([]string{
			strconv.FormatInt(c.CategoryID, 10),
			strconv.FormatInt(int64(c.PlatformID), 10),
			c.Day,
			strconv.FormatInt(c.MedianP, 10),
			strconv.FormatInt(c.P25P, 10),
			strconv.FormatInt(c.P75P, 10),
			fmt.Sprintf("%.2f", c.AvgDiscountPct),
			strconv.FormatInt(int64(c.SKUCount), 10),
		}); err != nil {
			return "", err
		}
	}
	w.Flush()
	return buf.String(), w.Error()
}
