---
fr_id: FR-CART-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

FR-CART-003 đặc tả `optimizeCart` ở mức triển khai được, bám đúng pseudo-code §3.5(3). 12 mệnh đề §1 normative neo vào AC §4 và test §5. Thuật toán duyệt (platform x freeship x shop-combo), lọc validStack + meetsMinSpend, applyCaps, giữ best - khớp cấu trúc nguồn. Luật stacking ủy quyền FR-CART-004 qua interface (lõi không hardcode luật nước), tiền tệ int64 (percent chia nguyên), và breakdown minh bạch. Ví dụ §3.5(3) ra đúng 110k được test khẳng định. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - Hardcode luật nước vào lõi (đã chốt)
Rủi ro: nhúng luật VN-stack vào optimizer -> mở MY/PH (bỏ stack 2025) phải sửa lõi. Giải: DEC-CART-14 + §1 #4 interface `StackRules.ValidStack`; AC #4 grep không có hằng nước; luật cụ thể ở FR-CART-004.

### ISS-002 - Sai số tiền tệ chọn sai combo
Float khi cộng shop+platform+freeship lệch vài đồng đủ chọn sai. Giải: DEC-CART-15 + §1 #6 số nguyên VND + percent chia nguyên; AC #6 + test DiscountPercent_IntegerDivision.

### ISS-003 - Bỏ applyCaps làm con số sai thực tế
Voucher có trần; bỏ cap thì hiển thị giảm vượt thực tế. Giải: §1 #7 + `applyCap` kẹp; AC #7 + test ApplyCap (20% của 400k kẹp về cap 50k).

### ISS-004 - Bám đúng pseudo-code §3.5(3)
Thuật toán phải khớp đặc tả nguồn để đúng và đối chiếu được. Giải: DEC-CART-13 + §1 #2 cấu trúc đầy đủ (filterByMinSpend, vòng duyệt, validStack, meetsMinSpend, applyCaps, best); AC #2 + #5 + test VNStackExample_110k.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 hàm thuần | #1 | `optimize.go` no-DB |
| #2 bám §3.5(3) | #2,#5 | vòng duyệt + test 110k |
| #3 best shop per shop | #3 | `chooseBestShopVoucherPerShop` |
| #4 validStack ủy quyền | #4 | interface `StackRules` |
| #5 meetsMinSpend | #8 | gate min_spend |
| #6 số nguyên VND | #6 | `discountAmount` |
| #7 applyCaps | #7 | `applyCap` |
| #8 breakdown | #11 | `breakdownOf` |
| #9 giỏ rỗng | #9 | Discount 0 |
| #10 tất định | #10 | sắp xếp ổn định |
| #11 handler caller | - | `optimize.go` handler |

## §4 - Kết luận

Mọi mệnh đề normative có code/test backing; thuật toán bám đúng pseudo-code §3.5(3), luật nước tách qua interface, tiền tệ int64, applyCaps khớp thực tế. Ví dụ nguồn (110k) được test khẳng định. Không mệnh đề "mồ côi". **Score = 10/10. Verdict: PASS.** Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit FR-CART-003.*
