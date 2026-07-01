/** @jest-environment jsdom */
import { readTiktokCart } from "../src/content/tiktok/cart-reader";
import { setConsent } from "../src/shared/consent";

const tiktokCartFixtureMain = `
  <div class="cart-list-wrapper">
    <div class="cart-item">
      <a href="/product/12345">Product 1</a>
      <span class="price-val">100,000 đ</span>
      <span class="qty-val">2</span>
    </div>
    <div class="cart-item">
      <div data-product-id="67890">Product 2</div>
      <div data-testid="item-price">50.000</div>
      <div data-testid="item-qty"><input value="1" /></div>
    </div>
  </div>
`;

describe("TikTok Cart Reader", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    
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

  test("đọc giỏ từ DOM TikTok đã render", async () => {
    await setConsent(["read_cart"]);
    document.body.innerHTML = tiktokCartFixtureMain;
    
    const msg = await readTiktokCart();
    console.log("MSG ITEMS:", msg.items);
    expect(msg.platform).toBe("tiktok");
    expect(msg.items.length).toBe(2);
    
    expect(msg.items[0].productId).toBe("12345");
    expect(msg.items[0].price).toBe(100000);
    expect(msg.items[0].qty).toBe(2);
    
    expect(msg.items[1].productId).toBe("67890");
    expect(msg.items[1].price).toBe(50000);
    expect(msg.items[1].qty).toBe(1);
  });
  
  test("trả rỗng nếu không có consent", async () => {
    await setConsent([]); // No read_cart
    document.body.innerHTML = tiktokCartFixtureMain;
    
    const msg = await readTiktokCart();
    expect(msg.items.length).toBe(0);
  });
  
  test("phát health signal khi DOM hỏng", async () => {
    await setConsent(["read_cart"]);
    document.body.innerHTML = `<div id="random">nothing here</div>`;
    
    const msg = await readTiktokCart();
    expect(msg.items.length).toBe(0);
    expect(chrome.runtime.sendMessage).toHaveBeenCalledWith(
      expect.objectContaining({ type: "HEALTH_SIGNAL", broke: "cart" })
    );
  });
});
