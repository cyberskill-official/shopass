/** @jest-environment jsdom */
import { readFileSync } from "fs";
import { join } from "path";
import { readLazadaCart } from "../src/content/lazada/cart-reader";
import { setConsent } from "../src/shared/consent";

const lazadaCartFixtureMain = `
<div class="cart-item">
  <a href="/product/abc-i12345.html">Product 1</a>
  <span class="current-price">100.000 ₫</span>
  <input class="quantity-input" value="2" />
</div>
<div class="cart-item">
  <div data-item-id="67890">Product 2</div>
  <div data-tracking="item-price">50.000 ₫</div>
  <div class="cart-item-qty">1</div>
</div>
`;

describe("Lazada No Akamai Evasion", () => {
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

  test("KHÔNG né/giả Akamai sensor, KHÔNG đọc cookie/mật khẩu", () => {
    for (const f of ["cart-reader", "dom-selectors", "index"]) {
      const src = readFileSync(join(__dirname, `../src/content/lazada/${f}.ts`), "utf8");
      expect(src).not.toMatch(/_abck|bm_sz|sensor_data/i);     // không sensor Akamai
      expect(src).not.toMatch(/document\.cookie/);             // không cookie
      expect(src).not.toMatch(/input\[type=["']?password/i);   // không mật khẩu
    }
  });

  test("payload Lazada KHÔNG chứa cookie/token/sensor", async () => {
    await setConsent(["read_cart"]);
    document.body.innerHTML = lazadaCartFixtureMain;
    const msg = await readLazadaCart();
    const flat = JSON.stringify(msg).toLowerCase();
    for (const b of ["cookie", "token", "_abck", "bm_sz", "sensor", "password"]) {
      expect(flat).not.toContain(b);
    }
  });
});
