package api

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strconv"

	"shopass/services/comply/internal/consent"
)

var errMissingUser = errors.New("missing user identity")

func userIDFromRequest(r *http.Request) (int64, error) {
	raw := r.Header.Get("X-User-Id")
	if raw == "" {
		return 0, errMissingUser
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errMissingUser
	}
	return id, nil
}

func requestMeta(r *http.Request) consent.ReqMeta {
	var ip *netip.Addr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if parsed, err := netip.ParseAddr(host); err == nil {
		ip = &parsed
	}

	var userAgent *string
	if raw := r.UserAgent(); raw != "" {
		userAgent = &raw
	}
	return consent.ReqMeta{IP: ip, UserAgent: userAgent}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeJSON(r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst) == nil
}
