package audit

import (
	"errors"
	"fmt"
	"strings"
)

var ErrForbiddenField = errors.New("forbidden field")

var forbiddenKeys = []string{"cookie", "token", "session", "authorization", "set-cookie"}

// GuardPayload tu choi payload chua truong nhay cam tu extension.
func GuardPayload(p map[string]any) error {
	for k := range p {
		lk := strings.ToLower(k)
		for _, f := range forbiddenKeys {
			if strings.Contains(lk, f) {
				return fmt.Errorf("%w: %s", ErrForbiddenField, k)
			}
		}
	}
	return nil
}
