package apns

import (
	"encoding/json"
	"errors"
	"fmt"
)

const maxPayloadBytes = 4096

// AlertPayload is the minimal aps alert shape for Shopass pushes.
type AlertPayload struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

type apsEnvelope struct {
	APS struct {
		Alert AlertPayload `json:"alert"`
		Sound string       `json:"sound,omitempty"`
	} `json:"aps"`
}

// BuildAlertPayload marshals an aps alert JSON and enforces the ~4KB APNs limit.
func BuildAlertPayload(title, body string) ([]byte, error) {
	var env apsEnvelope
	env.APS.Alert = AlertPayload{Title: title, Body: body}
	env.APS.Sound = "default"
	b, err := json.Marshal(env)
	if err != nil {
		return nil, err
	}
	if len(b) > maxPayloadBytes {
		return nil, fmt.Errorf("apns: payload %d exceeds %d bytes", len(b), maxPayloadBytes)
	}
	if title == "" && body == "" {
		return nil, errors.New("apns: empty alert")
	}
	return b, nil
}

// EnsurePayload returns job payload if valid size, else a default Shopass alert.
func EnsurePayload(raw json.RawMessage) ([]byte, error) {
	if len(raw) == 0 {
		return BuildAlertPayload("Shopass", "Bạn có thông báo mới")
	}
	if len(raw) > maxPayloadBytes {
		return nil, fmt.Errorf("apns: payload %d exceeds %d bytes", len(raw), maxPayloadBytes)
	}
	return []byte(raw), nil
}
