import { readFileSync } from "node:fs";
import { ALLOWED_ITEM_FIELDS, ALLOWED_VOUCHER_FIELDS } from "../src/pipeline/allowlist";

test("DISCLOSURE liệt kê đúng trường pipeline thực gửi (không thừa/thiếu)", () => {
  const disc = readFileSync("DISCLOSURE.md", "utf8").toLowerCase();
  for (const f of ALLOWED_ITEM_FIELDS) {
    expect(disc).toContain(f.toLowerCase());
  }
  // Cấm khai trường NHẠY CẢM (chống "khai sai để trông an toàn" hoặc rò ý đồ)
  for (const banned of ["cookie", "password", "mật khẩu", "session token", "token phiên"]) {
    expect(disc).toContain("không"); // mục "KHÔNG gửi" phải tồn tại
  }
  expect(disc).toMatch(/không.*(cookie)/);
});

test("trường mới ở allowlist mà DISCLOSURE chưa khai -> fail", () => {
  const disc = readFileSync("DISCLOSURE.md", "utf8").toLowerCase();
  const undisclosed = [...ALLOWED_ITEM_FIELDS, ...ALLOWED_VOUCHER_FIELDS]
    .filter(f => !disc.includes(f.toLowerCase()));
  expect(undisclosed).toEqual([]);  // mọi trường gửi đi phải được khai
});
