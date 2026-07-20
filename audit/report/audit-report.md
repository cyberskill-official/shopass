# SănDeal Security Audit

Version: 1.4.0 Commit: local

| Hook | Kết quả | Bằng chứng |
|---|---|---|
| Egress (động) | PASS | 3 outbound, 0 leaks |
| SBOM + vuln | PASS | 41 deps, 0 high CVE |
| Reproducible build | PASS | SHA-256 == dummy-hash |
| Payload guard (tĩnh) | PASS | 0 findings |

=> AUDIT PASS. Tự kiểm: `bash audit/run-security-audit.sh`
