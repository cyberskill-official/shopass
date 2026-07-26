export type ChangelogEntry = {
  version: string;
  date: string;
  title: string;
  items: string[];
};

/** Newest first. Hook deploy notes from R15 here when CI publishes releases. */
export const CHANGELOG: ChangelogEntry[] = [
  {
    version: "0.3.0",
    date: "2026-07-26",
    title: "Growth surface — pricing, tools, onboarding, compare",
    items: [
      "Bảng giá + Premium waitlist (R39)",
      "Kiểm tra sale ảo + lịch sale (R43)",
      "Onboarding tới cảnh báo đầu tiên (R45)",
      "So sánh BeeCost + trang thay Honey (R44)",
      "Blog / changelog / RSS (R47)",
    ],
  },
  {
    version: "0.2.0",
    date: "2026-07-26",
    title: "Trust + landing",
    items: [
      "Landing công khai (R38)",
      "Chính sách / điều khoản / minh bạch (R34–R35)",
      "ML model_run gate (R26)",
    ],
  },
];
