import { chromium } from "playwright";
import { LazadaAdapter } from "../adapter";
import { ChallengedError } from "../../../render";
import * as http from "http";

describe("Lazada Adapter", () => {
  let server: http.Server;
  let port: number;

  beforeAll((done) => {
    server = http.createServer((req, res) => {
      if (req.url === "/challenge") {
        res.writeHead(200, { "Content-Type": "text/html" });
        res.end("<html><head><title>Access Denied</title></head><body>akamai access denied</body></html>");
      } else if (req.url === "/pdp") {
        res.writeHead(200, { "Content-Type": "text/html" });
        res.end(`
          <html>
            <body>
              <script>
                window.pageData = {"core":{"price":"120000","originPrice":"180000"},"flashSale":true};
              </script>
              <div class="pdp-price--current">₫120.000</div>
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

  it("chặn akamai verify/challenge", async () => {
    const browser = await chromium.launch();
    const ctx = await browser.newContext();
    const adapter = new LazadaAdapter();
    
    await expect(adapter.fetch(ctx, { url: `http://localhost:${port}/challenge`, productId: "1" }))
      .rejects.toThrow(ChallengedError);
      
    await browser.close();
  });

  it("đọc thành công giá json nhúng", async () => {
    const browser = await chromium.launch();
    const ctx = await browser.newContext();
    const adapter = new LazadaAdapter();
    
    const snap = await adapter.fetch(ctx, { url: `http://localhost:${port}/pdp`, productId: "laz123" });
    
    expect(snap.productId).toBe("laz123");
    expect(snap.price).toBe(120000);
    expect(snap.listPrice).toBe(180000);
    expect(snap.flashSale).toBe(true);
      
    await browser.close();
  });
});
