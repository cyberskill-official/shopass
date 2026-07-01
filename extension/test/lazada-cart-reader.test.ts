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

describe("Lazada Cart Reader", () => {
  beforeEach(() => {
    let storageData: any = {};
    global.chrome = {
      runtime: { sendMessage: jest.fn() },
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

  test("đọc giỏ từ DOM Lazada đã render", async () => {
    await setConsent(["read_cart"]);
    document.body.innerHTML = lazadaCartFixtureMain;          // 2 item
    const msg = await readLazadaCart();
    expect(msg.platform).toBe("lazada");
    expect(msg.items.length).toBe(2);
    expect(msg.items[0].productId).toBe("12345");
    expect(msg.items[0].price).toBe(100000);
    expect(msg.items[0].qty).toBe(2);
    
    expect(msg.items[1].productId).toBe("67890");
    expect(msg.items[1].price).toBe(50000);
    expect(msg.items[1].qty).toBe(1);
  });

  test("reader tái dùng normalize/health FR-EXT-002", () => {
    const src = readFileSync(join(__dirname, "../src/content/lazada/cart-reader.ts"), "utf8");
    expect(src).toMatch(/from ["']\.\.\/\.\.\/shared\/normalize["']/);
    expect(src).toMatch(/from ["']\.\.\/\.\.\/shared\/health["']/);
  });

  test("Akamai challenge / giỏ rỗng → rỗng lịch sự, không vượt challenge", async () => {
    await setConsent(["read_cart"]);
    document.body.innerHTML = "<div id='akamai-challenge'>...</div>";
    const msg = await readLazadaCart();
    expect(msg.items).toEqual([]);                            // không thử vượt
  });
});
