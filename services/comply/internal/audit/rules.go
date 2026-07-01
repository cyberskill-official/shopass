package audit

import "regexp"

type Rule struct {
	Name    string
	Pattern *regexp.Regexp
	Hint    string
}

// bannedRules contains the list of rules to enforce no-cleartext and token-not-on-server invariants.
var bannedRules = []Rule{
	{
		Name:    "cleartext_password",
		Pattern: regexp.MustCompile(`(?i)\b(plain_?password|password)\s+TEXT\b`),
		Hint:    "Mat khau phai la pwd_hash argon2id (FR-AUTH-001)",
	},
	{
		Name:    "platform_session_token",
		Pattern: regexp.MustCompile(`(?i)(shopee|tiktok|lazada)_?(token|cookie|session)|platform_(access_)?token`),
		Hint:    "Token phien san KHONG duoc co o backend (token-not-on-server)",
	},
	{
		Name:    "hardcoded_secret",
		Pattern: regexp.MustCompile(`(?i)(api_?key|db_?password)\s*[:=]\s*["'][A-Za-z0-9/+]{12,}["']`),
		Hint:    "Secret phai doc tu Vault (FR-INFRA-003)",
	},
	{
		Name:    "weak_password_hash",
		Pattern: regexp.MustCompile(`(?i)\b(md5|sha1)\s*\(`),
		Hint:    "Cam bam yeu cho credential; dung argon2id",
	},
}
