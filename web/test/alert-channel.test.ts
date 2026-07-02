import { validateAlert, CHANNELS } from "../lib/alerts/validate";

describe("Alert channels validation", () => {
  it("channel chỉ {push,email,sms}; rỗng hoặc lạ báo lỗi", () => {
    expect(CHANNELS).toEqual(["push", "email", "sms"]);
    expect(validateAlert("real_sale", null, [])).toMatch(/ít nhất một kênh/i);
    expect(validateAlert("real_sale", null, ["zalo" as any])).toMatch(/không hợp lệ/i);
    expect(validateAlert("real_sale", null, ["push", "email"])).toBeNull();
  });
});
