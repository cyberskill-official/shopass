import { extractTikTok } from "../extract";
import { PageView } from "../../types";

describe("TikTok Extract", () => {
  it("ưu tiên đọc từ JSON nhúng khi có ProductModule", () => {
    const page: PageView = {
      embeddedJSON: (sel) => JSON.stringify({
        ProductModule: { price: "99000", originalPrice: "150000", isFlashSale: true }
      }),
      text: (sel) => "100", // DOM giả
      exists: (sel) => false,
    };
    
    const res = extractTikTok(page);
    expect(res.source).toBe("json");
    expect(res.price).toBe(99000);
    expect(res.listPrice).toBe(150000);
    expect(res.flashSale).toBe(true);
  });

  it("đọc từ JSON nhúng dạng ItemModule", () => {
    const page: PageView = {
      embeddedJSON: (sel) => JSON.stringify({
        ItemModule: {
          "123": {
            price: { priceValue: "55000", originalPriceValue: "80000" },
            activityInfo: { isFlashSale: false }
          }
        }
      }),
      text: (sel) => "100", // DOM giả
      exists: (sel) => false,
    };
    
    const res = extractTikTok(page);
    expect(res.source).toBe("json");
    expect(res.price).toBe(55000);
    expect(res.listPrice).toBe(80000);
    expect(res.flashSale).toBe(false);
  });

  it("fallback đọc DOM text khi JSON không hợp lệ", () => {
    const page: PageView = {
      embeddedJSON: (sel) => null,
      text: (sel) => {
        if (sel.includes("origin")) return "₫150.000";
        return "₫99.000";
      },
      exists: (sel) => true, // flash
    };
    
    const res = extractTikTok(page);
    expect(res.source).toBe("dom");
    expect(res.price).toBe(99000);
    expect(res.listPrice).toBe(150000);
    expect(res.flashSale).toBe(true);
  });

  it("ném lỗi nếu parse DOM cũng hỏng", () => {
    const page: PageView = {
      embeddedJSON: (sel) => null,
      text: (sel) => null,
      exists: (sel) => false,
    };
    
    expect(() => extractTikTok(page)).toThrow(/extract fail/);
  });
});
