/** @jest-environment jsdom */
import fs from "fs";
import path from "path";
import { readTiktokCart } from "../src/content/tiktok/cart-reader";
import { setConsent } from "../src/shared/consent";

describe("TikTok No API Sign", () => {
  beforeEach(() => {
    let storageData: any = {};
    global.chrome = {
      storage: {
        local: {
          get: jest.fn(async (key) => {
            if (typeof key === "string") return { [key]: storageData[key] };
            return storageData;
          }),
          set: jest.fn(async (data) => {
            storageData = { ...storageData, ...data };
          })
        }
      }
    } as any;
  });

  test("KHÔNG ký/gửi msToken/_signature/X-Bogus, KHÔNG đọc cookie", () => {
    const files = ["cart-reader.ts", "dom-selectors.ts", "spa-observer.ts", "index.ts"];
    for (const f of files) {
      const src = fs.readFileSync(path.join(__dirname, "../src/content/tiktok", f), "utf8");
      expect(src).not.toMatch(/msToken|_signature|X-Bogus/i);
      expect(src).not.toMatch(/document\.cookie/);
      expect(src).not.toMatch(/input\[type=["']?password/i);
    }
  });

  test("payload TikTok KHÔNG chứa cookie/token", async () => {
    await setConsent(["read_cart"]);
    document.body.innerHTML = `<div class="cart-list-wrapper"><div class="cart-item"><div data-product-id="1"><span class="price-val">10</span><span class="qty-val">1</span></div></div></div>`;
    
    const msg = await readTiktokCart();
    const flat = JSON.stringify(msg).toLowerCase();
    
    for (const b of ["cookie", "token", "mstoken", "signature", "x-bogus", "password"]) {
      expect(flat).not.toContain(b);
    }
  });
});
