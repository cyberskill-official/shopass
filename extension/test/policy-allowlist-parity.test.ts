import { COLLECTED_FIELDS, NEVER_COLLECTED } from "../src/policy/collected-fields";
import { ALLOWED_ITEM_FIELDS } from "../src/pipeline/allowlist";

describe("Policy <-> Allowlist Parity", () => {
  it("mọi trường item ở allowlist đều được mô tả trong chính sách", () => {
    const described = new Set(COLLECTED_FIELDS.map(f => f.field));
    for (const f of ALLOWED_ITEM_FIELDS) {
      expect(described.has(f)).toBe(true);
    }
  });

  it("chính sách KHÔNG mô tả trường nào pipeline không gửi (không phóng đại/giấu)", () => {
    const allowed = new Set<string>(["platform", ...ALLOWED_ITEM_FIELDS]);
    for (const f of COLLECTED_FIELDS) {
      expect(allowed.has(f.field)).toBe(true);
    }
  });

  it("mọi trường thu thập đều có mục đích + cơ sở pháp lý (không 'thu để dành')", () => {
    for (const f of COLLECTED_FIELDS) {
      expect(f.purpose.length).toBeGreaterThan(0);
      expect(f.legalBasis.length).toBeGreaterThan(0);
    }
  });

  it("danh sách NEVER_COLLECTED bao gồm các loại nhạy cảm cốt lõi", () => {
    for (const must of ["cookie", "mật khẩu", "token phiên sàn"]) {
      expect(NEVER_COLLECTED).toContain(must as any);
    }
  });
});
