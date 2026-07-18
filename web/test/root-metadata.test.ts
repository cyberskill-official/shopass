import { metadata } from "../app/layout";

describe("root metadata", () => {
  it("uses the Shopass brand instead of scaffold metadata", () => {
    expect(metadata.description).toMatch(/Theo dõi lịch sử giá/i);
    expect(metadata.title).not.toBe("Create Next App");
  });
});
