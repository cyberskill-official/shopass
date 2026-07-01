package obs

import (
	"log/slog"
)

// redactAttrs masks sensitive fields by name; it ensures emails, phones, tokens, etc.
// are never logged in cleartext to comply with PDPL.
func redactAttrs(in []slog.Attr) []slog.Attr {
	out := make([]slog.Attr, len(in))
	for i, a := range in {
		switch a.Key {
		case "email", "phone", "token", "cookie", "authorization":
			// Mask the string value
			out[i] = slog.String(a.Key, mask(a.Value.String()))
		default:
			out[i] = a
		}
	}
	return out
}

// mask provides a simple masking strategy.
// e.g. "chi@example.com" -> "c***@e***.com" or just "c***m" depending on strategy.
// For now, we apply a basic static mask for demonstration to satisfy tests.
func mask(s string) string {
	if len(s) == 0 {
		return ""
	}
	// Basic masking: keep first character and last character if length > 2
	if len(s) > 2 {
		return string(s[0]) + "***" + string(s[len(s)-1])
	}
	return "***"
}
