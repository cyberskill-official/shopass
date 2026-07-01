/** @jest-environment jsdom */
import { onCartRendered } from "../src/content/tiktok/spa-observer";

describe("TikTok SPA Observer", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
  });

  test("kích hoạt ngay nếu giỏ đã render sẵn", () => {
    document.body.innerHTML = `<div class="cart-list-wrapper"></div>`;
    const cb = jest.fn();
    
    const disconnect = onCartRendered(cb);
    expect(cb).toHaveBeenCalledTimes(1);
    
    disconnect();
  });

  test("chờ mutation rồi mới kích hoạt khi giỏ xuất hiện", async () => {
    const cb = jest.fn();
    const disconnect = onCartRendered(cb);
    
    expect(cb).not.toHaveBeenCalled();
    
    // Simulate SPA route change bringing in cart
    document.body.innerHTML = `<div class="cart-list-wrapper"></div>`;
    
    // Give observer a tick
    await new Promise(r => setTimeout(r, 10));
    
    expect(cb).toHaveBeenCalledTimes(1);
    disconnect();
  });
});
