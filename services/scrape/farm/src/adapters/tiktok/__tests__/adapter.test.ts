import { chromium } from "playwright";
import { TikTokAdapter } from "../adapter";
import { ChallengedError } from "../../../render";
import * as http from "http";

describe("TikTok Adapter", () => {
  let server: http.Server;
  let port: number;

  beforeAll((done) => {
    server = http.createServer((req, res) => {
      if (req.url === "/challenge") {
        res.writeHead(200, { "Content-Type": "text/html" });
        res.end("<html><head><title>Verify to continue</title></head><body>Verify you are human</body></html>");
      } else if (req.url === "/pdp") {
        res.writeHead(200, { "Content-Type": "text/html" });
        res.end(`
          <html>
            <body>
              <script id="__NEXT_DATA__">
                {"props":{"pageProps":{"initialData":{"productInfo":{"price":"45000","originalPrice":"60000","isFlashSale":true}}}}}
              </script>
              <div data-e2e="product-price">₫45.000</div>
            </body>
          </html>
        `);
      } else {
        res.writeHead(404);
        res.end();
      }
    });
    server.listen(0, () => {
      port = (server.address() as any).port;
      done();
    });
  });

  afterAll((done) => {
    server.close(done);
  });

  it("chặn verify/challenge", async () => {
    const browser = await chromium.launch();
    const ctx = await browser.newContext();
    const adapter = new TikTokAdapter();
    
    await expect(adapter.fetch(ctx, { url: `http://localhost:${port}/challenge`, productId: "1" }))
      .rejects.toThrow(ChallengedError);
      
    await browser.close();
  });

  it("đọc thành công giá json nhúng", async () => {
    const browser = await chromium.launch();
    const ctx = await browser.newContext();
    const adapter = new TikTokAdapter();
    
    const snap = await adapter.fetch(ctx, { url: `http://localhost:${port}/pdp`, productId: "tiktok123" });
    
    expect(snap.productId).toBe("tiktok123");
    expect(snap.price).toBe(45000);
    expect(snap.listPrice).toBe(60000);
    expect(snap.flashSale).toBe(true);
      
    await browser.close();
  });
});
