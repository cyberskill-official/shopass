/** @jest-environment jsdom */
import { readCart } from "../src/content/shopee/cart-reader";
import * as fs from "fs/promises";
import * as path from "path";

test("payload KHÔNG chứa cookie/token", async () => {
  // We need to mock fetch to return some dummy data so readCart works
  global.fetch = jest.fn().mockResolvedValue({
    ok: true,
    json: async () => ({ data: { shop_orders: [{ items: [{ itemid: 123, price: 10000000, amount: 1 }] }] } })
  });
  
  // mock DOM
  document.body.innerHTML = `
    <div class="shopee-voucher">
      <div class="voucher-code">FREESHIP</div>
    </div>
  `;

  const msg = await readCart();
  const flat = JSON.stringify(msg).toLowerCase();
  for (const banned of ["cookie", "token", "session", "authorization", "password"]) {
    expect(flat).not.toContain(banned);
  }
});

test("mã nguồn KHÔNG đọc document.cookie / password input", async () => {
  const dir = path.join(__dirname, "../src/content/shopee");
  for (const f of ["cart-reader", "voucher-reader", "api-client", "dom-selectors", "index"]) {
    const src = await fs.readFile(path.join(dir, `${f}.ts`), "utf8");
    expect(src).not.toMatch(/document\.cookie/);
    expect(src).not.toMatch(/Set-Cookie|Authorization/i);
    expect(src).not.toMatch(/input\[type=["']?password/i);
  }
});
