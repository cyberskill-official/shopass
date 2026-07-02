import { extractLazada } from "../extract";
import { PageView } from "../../types";

describe("Lazada Extract", () => {
  it("ưu tiên đọc từ window.pageData JSON nhúng", () => {
    const page: PageView = {
      embeddedJSON: (sel) => `window.pageData = {"core":{"price":"120000","originPrice":"180000"},"flashSale":true};`,
      text: (sel) => "100", 
      exists: (sel) => false,
    };
    
    const res = extractLazada(page);
    expect(res.source).toBe("json");
    expect(res.price).toBe(120000);
    expect(res.listPrice).toBe(180000);
    expect(res.flashSale).toBe(true);
  });

  it("đọc từ JSON nhúng khi cấu trúc là __moduleData__ có block price", () => {
    const page: PageView = {
      embeddedJSON: (sel) => `window.__moduleData__ = {"price":{"current":"75000","original":"90000"},"activity":{"type":"flash_sale"}};`,
      text: (sel) => "100", 
      exists: (sel) => false,
    };
    
    const res = extractLazada(page);
    expect(res.source).toBe("json");
    expect(res.price).toBe(75000);
    expect(res.listPrice).toBe(90000);
    expect(res.flashSale).toBe(true);
  });

  it("fallback đọc DOM text khi JSON không hợp lệ", () => {
    const page: PageView = {
      embeddedJSON: (sel) => null,
      text: (sel) => {
        if (sel.includes("deleted")) return "₫150.000";
        return "₫99.000";
      },
      exists: (sel) => true,
    };
    
    const res = extractLazada(page);
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
    
    expect(() => extractLazada(page)).toThrow(/extract fail/);
  });
});
