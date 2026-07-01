package canon

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// CanonicalKey dựng key xác định (DEC-PRICE-21): brand + model + hash thuộc tính.
// Cùng input → cùng output (sắp xếp thuộc tính trước khi hash).
func CanonicalKey(brand, model string, attrs map[string]string) string {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys) // xác định, không phụ thuộc thứ tự map
	
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(attrs[k])
		b.WriteByte(';')
	}
	
	sum := sha256.Sum256([]byte(b.String()))
	return brand + ":" + model + ":" + hex.EncodeToString(sum[:])[:12]
}
