# SănDeal Extension - Disclosure dữ liệu

## Dữ liệu extension GỬI về server (đúng tập tối thiểu)
- platform: sàn đang xem (shopee | tiktok | lazada)
- productId: ID sản phẩm công khai
- price: giá hiển thị (số nguyên VND)
- qty: số lượng trong giỏ
- voucher hiển thị: code (mã voucher công khai), minSpend (mức chi tối thiểu), discountText (mô tả giảm giá)

## Dữ liệu extension KHÔNG BAO GIỜ gửi
- KHÔNG cookie phiên (Shopee/TikTok/Lazada)
- KHÔNG mật khẩu
- KHÔNG token phiên sàn / header xác thực
- KHÔNG email / số điện thoại / tên / địa chỉ

## Vì sao cần từng quyền
- host_permissions (shopee.vn, tiktok.com, lazada.vn): đọc giá/giỏ trên đúng trang sàn
- storage: lưu cấu hình + state cục bộ (service worker ephemeral, TASK-EXT-001)
- alarms: lập lịch quét định kỳ (>=30s) thay cho setInterval

Mã nguồn: <https://github.com/shopass/sandeal-extension> Xem chi tiết [Chính sách Tối thiểu hóa Dữ liệu](../docs/trust/DATA-MINIMIZATION-POLICY.md)
