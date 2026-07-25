# SănDeal Security Audit

Version: 1.4.0
Commit: local

| Hook | Kết quả | Bằng chứng |
|---|---|---|
| Egress (động) | PASS | jest egress suite exited 0 |
| SBOM + vuln | FAIL / NOT_RUN | SBOM generation and/or vuln scan not implemented or failed |
| Reproducible build | FAIL / NOT_RUN | hook not implemented — fail closed |
| Payload guard (tĩnh) | PASS | go test ./internal/audit/... exited 0 |

=> AUDIT FAIL. Hooks that are not implemented fail closed. Re-run: `bash audit/run-security-audit.sh`
