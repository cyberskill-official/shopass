# SănDeal Third-Party Audit Guide

Tài liệu này hướng dẫn bạn cách tự chạy lại bộ kiểm tra bảo mật (security audit) của SănDeal trên máy cá nhân để chứng minh SănDeal **KHÔNG** gửi cookie, mật khẩu, hay thông tin định danh (PII) về server.

## 1. Yêu cầu hệ thống
- Node.js >= 18
- npm hoặc pnpm
- Playwright (tự động cài khi chạy npm install)
- Git

## 2. Các bước thực hiện

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
```

**Bước 3: Chạy bộ Security Audit**
```bash
bash audit/run-security-audit.sh
```

## 3. Kết quả mong đợi

Bộ audit bao gồm 4 phần:
1. **Egress test (động)**: Mở trình duyệt giả lập, tiêm cookie giả (mô phỏng đã đăng nhập), bắt extension quét giỏ hàng, và chặn ở tầng mạng để đảm bảo KHÔNG CÓ cookie nào "rời máy".
2. **SBOM + vuln scan**: Quét lỗ hổng (CVE) trên toàn bộ dependencies của extension.
3. **Verify reproducible build**: Đảm bảo mã bạn tải từ cửa hàng tiện ích (Chrome Store) khớp từng byte (SHA-256) với mã nguồn công khai.
4. **Payload guard (tĩnh)**: Trình quét Go backend từ chối bất kỳ payload nào chứa dấu hiệu credential.

Khi chạy hoàn tất, script sẽ báo cáo `AUDIT PASS`.

## 4. Kiểm tra báo cáo

Sau khi chạy xong, kết quả chi tiết sẽ được xuất ra tại:
- `audit/report/audit-report.md` (định dạng Markdown dễ đọc)
- `audit/report/audit-report.json` (định dạng máy đọc)

## 5. Thử nghiệm "Negative Control" (Chứng minh công cụ hoạt động)

Để chắc chắn công cụ thực sự đang kiểm tra nghiêm ngặt, bạn có thể cố ý làm công cụ thất bại bằng cách thêm dòng mã rò rỉ cookie trong `audit/egress/egress-guard.test.ts`. Khi bạn chạy lại `bash audit/run-security-audit.sh`, công cụ phải báo `FAIL`.
