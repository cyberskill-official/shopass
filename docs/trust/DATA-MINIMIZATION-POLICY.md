# Chính sách Tối thiểu hóa Dữ liệu (Data Minimization Policy)

*SănDeal cam kết tuân thủ nguyên tắc Tối thiểu hóa dữ liệu của Luật Bảo vệ Dữ liệu Cá nhân (Luật 91/2025).*

## Nguyên tắc cốt lõi: Xử lý Local-First

SănDeal áp dụng kiến trúc **local-first**. Điều này có nghĩa là mọi quá trình chuẩn hóa, khử trùng, và chọn lọc dữ liệu đều diễn ra **ngay trên thiết bị của bạn** (trong trình duyệt). Máy chủ của chúng tôi (Backend) chỉ nhận phần dữ liệu đã được làm sạch hoàn toàn. Chúng tôi không thu thập dữ liệu thô để "lọc giúp" ở máy chủ.

## Chúng tôi KHÔNG BAO GIỜ thu thập
Chúng tôi cam kết tường minh KHÔNG thu thập hoặc gửi đi bất kỳ dữ liệu nhạy cảm nào sau đây:
- Cookie đăng nhập
- Mật khẩu
- Token phiên sàn (Session Token)
- Header xác thực
- Email
- Số điện thoại
- Tên người dùng
- Địa chỉ
- Định danh người dùng sàn thật

Bạn có thể tự kiểm chứng điều này qua mã nguồn mở của chúng tôi (xem `DISCLOSURE.md`).

## Chúng tôi thu thập gì và vì sao

SănDeal chỉ thu thập chính xác những trường dữ liệu tối thiểu sau đây để có thể cung cấp dịch vụ cho bạn:

| Dữ liệu | Vì sao cần | Cơ sở pháp lý (PDPL) |
|---|---|---|
| Sàn đang xem (`platform`) | Tra đúng dữ liệu giá sàn đó | Đồng thuận - mục đích theo dõi giá |
| ID sản phẩm (`productId`) | Hiện lịch sử giá + cảnh báo sale ảo | Đồng thuận - mục đích theo dõi giá |
| Giá hiển thị (`price`) | Theo dõi biến động giá để cảnh báo và vẽ biểu đồ | Đồng thuận - mục đích theo dõi giá |
| Số lượng trong giỏ (`qty`) | Tính tối ưu voucher/giỏ hàng cho bạn | Đồng thuận - mục đích tối ưu giỏ |

Tất cả các trường khác đều bị loại bỏ ngay tại thiết bị của bạn trước khi gửi về máy chủ.

## Chính sách lưu trữ (Retention)

Dữ liệu giá thu thập được lưu trữ theo chính sách truy cập/xóa (DSAR) nhất quán. Lịch sử giá (price snapshot) có thể được lưu trữ ẩn danh để phục vụ cảnh báo biến động. Dữ liệu sẽ được bảo vệ ẩn danh và khi chia sẻ dữ liệu tổng hợp (nếu có, cho B2B) sẽ tuân thủ nghiêm ngặt điều kiện k-anonymity, đảm bảo không thể truy vết ngược lại cá nhân bạn.
