package notif

import (
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"html"
)

type Rendered struct {
	Title string
	Body  string
}

// printer is used for formatting numbers with dots for thousands
var printer = message.NewPrinter(language.Vietnamese)

func formatVND(v int64) string {
	return printer.Sprintf("%d VND", v)
}

func intField(d map[string]any, key string) (int64, bool) {
	v, ok := d[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case int64:
		return val, true
	case int:
		return int64(val), true
	case float64:
		return int64(val), true
	}
	return 0, false
}

func floatField(d map[string]any, key string) (float64, bool) {
	v, ok := d[key]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int64:
		return float64(val), true
	case int:
		return float64(val), true
	}
	return 0, false
}

func strField(d map[string]any, key string) (string, bool) {
	v, ok := d[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

var templates = map[string]func(map[string]any) (Rendered, error){
	"price_below": func(d map[string]any) (Rendered, error) {
		price, ok := intField(d, "price")
		if !ok {
			return Rendered{}, fmt.Errorf("template price_below thiếu price")
		}

		// Optional fields, but we escape if present
		productName, _ := strField(d, "product_name")
		productName = html.EscapeString(productName)

		body := fmt.Sprintf("Sản phẩm bạn theo dõi còn %s.", formatVND(price))
		if productName != "" {
			body = fmt.Sprintf("Sản phẩm %s bạn theo dõi còn %s.", productName, formatVND(price))
		}

		return Rendered{
			Title: "Giá đã giảm về mức bạn chờ",
			Body:  body,
		}, nil
	},
	"drop_pct": func(d map[string]any) (Rendered, error) {
		pct, ok := intField(d, "pct")
		if !ok {
			return Rendered{}, fmt.Errorf("template drop_pct thiếu pct")
		}
		return Rendered{
			Title: "Giá giảm sâu",
			Body:  fmt.Sprintf("Sản phẩm vừa giảm %d%%.", pct),
		}, nil
	},
	"real_sale": func(d map[string]any) (Rendered, error) {
		price, ok := intField(d, "price")
		if !ok {
			return Rendered{}, fmt.Errorf("template real_sale thiếu price")
		}
		return Rendered{
			Title: "Khuyến mãi thực chất",
			Body:  fmt.Sprintf("Sản phẩm đang sale thật sự, giá còn %s.", formatVND(price)),
		}, nil
	},
	"bottom_predicted": func(d map[string]any) (Rendered, error) {
		if price, ok := intField(d, "price"); ok {
			return Rendered{
				Title: "Giá chạm đáy dự kiến",
				Body:  fmt.Sprintf("Đây là thời điểm tốt nhất để mua, giá %s.", formatVND(price)),
			}, nil
		}
		// dealsvc nightly batch sends p_bottom_14d without a spot price.
		if p, ok := floatField(d, "p_bottom_14d"); ok {
			pct := int(p * 100)
			if p > 0 && p <= 1 && pct == 0 {
				pct = 1
			}
			return Rendered{
				Title: "Giá chạm đáy dự kiến",
				Body:  fmt.Sprintf("Xác suất chạm đáy trong 14 ngày khoảng %d%% — đây là thời điểm tốt để mua.", pct),
			}, nil
		}
		return Rendered{}, fmt.Errorf("template bottom_predicted thiếu price hoặc p_bottom_14d")
	},
}

func Render(template string, data map[string]any) (Rendered, error) {
	fn, ok := templates[template]
	if !ok {
		return Rendered{}, fmt.Errorf("template không tồn tại: %s", template)
	}
	return fn(data)
}
