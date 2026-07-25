# SănDeal Third-Party Audit Guide

Tài liệu này hướng dẫn bạn cách tự chạy lại bộ kiểm tra bảo mật (security audit) của SănDeal trên máy cá nhân để chứng minh SănDeal **KHÔNG** gửi cookie, mật khẩu, hay thông tin định danh (PII) về server.

## 1. Yêu cầu hệ thống
- Node.js >= 18
- npm hoặc pnpm
- Playwright (tự động cài khi chạy npm install)
- Git

## 2. Trạng thái hiện tại (fail-closed)

Bốn hook audit **chưa được nối đầy đủ**. `bash audit/run-security-audit.sh` **fail closed**: nó thoát mã khác 0 và ghi `AUDIT FAIL` / `NOT_RUN` cho mọi hook chưa có bằng chứng thật. Không tin bất kỳ báo cáo `AUDIT PASS` cũ nào trong lịch sử git trước khi các hook được triển khai thật.

Hook còn thiếu:
1. **Egress test** — cần jest suite chạy được trong `audit/`.
2. **SBOM + vuln scan** — cần tool CycloneDX / scanner thật (hiện `generate-sbom.sh` và `scan-vulnerabilities.sh` cố ý `exit 1`).
3. **Reproducible build** — cần `extension/scripts/verify-reproducible.sh` + artifact `SHIPPED`.
4. **Payload guard** — cần `services/comply/internal/audit` tests.

## 3. Các bước thực hiện (khi hooks đã wired)

**Bước 1: Clone mã nguồn SănDeal**
```bash
git clone https://github.com/shopass/sandeal-extension.git
cd sandeal-extension
```

**Bước 2: Cài đặt dependencies**
```bash
cd extension
npm install
cd ..
cd audit && npm install && cd ..
```

**Bước 3: Chạy bộ Security Audit**
```bash
bash audit/run-security-audit.sh
```

Kỳ vọng hôm nay: `AUDIT FAIL`. Chỉ khi cả bốn hook xanh mới được phép báo `AUDIT PASS`.

## 4. Kiểm tra báo cáo

Sau khi chạy xong, kết quả chi tiết sẽ được xuất ra tại:
- `audit/report/audit-report.md` (định dạng Markdown dễ đọc)
- `audit/report/audit-report.json` (định dạng máy đọc)

Verdict trong các file này phải khớp exit code của runner — không bao giờ hardcode `PASS`.

## 5. Thử nghiệm "Negative Control"

Khi egress hook đã wired, cố ý thêm dòng mã rò rỉ cookie trong `audit/egress/egress-guard.test.ts`. Chạy lại `bash audit/run-security-audit.sh` phải vẫn báo `FAIL`.
