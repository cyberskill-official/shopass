import PrivacyPolicyPage, {
  metadata as privacyMetadata,
} from "../app/(marketing)/chinh-sach-bao-mat/page";
import TermsPage, { metadata as termsMetadata } from "../app/(marketing)/dieu-khoan/page";

describe("R34 legal pages", () => {
  it("privacy page has VN metadata and draft-facing title", () => {
    expect(privacyMetadata.title).toMatch(/bảo mật/i);
    expect(PrivacyPolicyPage()).toBeTruthy();
  });

  it("terms page has VN metadata", () => {
    expect(termsMetadata.title).toMatch(/Điều khoản/i);
    expect(TermsPage()).toBeTruthy();
  });
});
