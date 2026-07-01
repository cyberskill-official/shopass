import { openRealtime, closeRealtime, isWsOpen } from "../src/sync/ws-client";
import * as fs from "fs";
import * as path from "path";

describe("ws-lifecycle", () => {
  beforeEach(() => {
    // mock WebSocket
    (global as any).WebSocket = class {
      readyState = 1;
      onclose: (() => void) | null = null;
      close() {
        this.readyState = 3; // CLOSED
        if (this.onclose) this.onclose();
      }
    };
  });

  afterEach(() => {
    closeRealtime();
    delete (global as any).WebSocket;
  });

  it("WSS is opened on demand and closed when done", () => {
    expect(isWsOpen()).toBe(false);
    openRealtime();
    expect(isWsOpen()).toBe(true);
    closeRealtime();
    expect(isWsOpen()).toBe(false);
  });

  it("service worker does not open WSS globally", () => {
    const swPath = path.join(__dirname, "..", "src", "background", "service-worker.ts");
    if (fs.existsSync(swPath)) {
      const src = fs.readFileSync(swPath, "utf-8");
      // Mở ws ở top level thường có kiểu `new WebSocket(`
      expect(src).not.toMatch(/^new WebSocket/m); // wait this regex is a bit specific, let's just use the exact string
      expect(src).not.toMatch(/new WebSocket/g); 
    }
  });
});
