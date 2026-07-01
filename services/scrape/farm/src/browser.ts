import { Browser, BrowserContext } from 'playwright';
import { FingerprintProfile } from './fingerprint';

export interface ProxySession {
  url: string;
  user?: string;
  pass?: string;
}

export function spoofScript(p: FingerprintProfile): string {
  return `
    Object.defineProperty(navigator, 'webdriver', { get: () => false });
    
    // Spoof Canvas
    const originalGetContext = HTMLCanvasElement.prototype.getContext;
    HTMLCanvasElement.prototype.getContext = function(type, ...args) {
      const ctx = originalGetContext.call(this, type, ...args);
      if (type === '2d' && ctx) {
        const originalGetImageData = ctx.getImageData;
        ctx.getImageData = function(...args2) {
          const imageData = originalGetImageData.apply(this, args2);
          // add stable noise based on seed
          const seed = ${p.canvasNoiseSeed};
          if (imageData && imageData.data) {
            for (let i = 0; i < imageData.data.length; i += 4) {
              imageData.data[i] = (imageData.data[i] + seed) % 256;
            }
          }
          return imageData;
        };
      }
      return ctx;
    };

    // Spoof WebGL
    const originalGetParameter = WebGLRenderingContext.prototype.getParameter;
    WebGLRenderingContext.prototype.getParameter = function(parameter) {
      if (parameter === 37445) return '${p.webgl.vendor}';
      if (parameter === 37446) return '${p.webgl.renderer}';
      return originalGetParameter.call(this, parameter);
    };
    
    const originalGetParameter2 = WebGL2RenderingContext.prototype.getParameter;
    WebGL2RenderingContext.prototype.getParameter = function(parameter) {
      if (parameter === 37445) return '${p.webgl.vendor}';
      if (parameter === 37446) return '${p.webgl.renderer}';
      return originalGetParameter2.call(this, parameter);
    };
  `;
}

export async function newPatchedContext(
  browser: Browser, p: FingerprintProfile, proxy?: ProxySession,
): Promise<BrowserContext> {
  const options: any = {
    userAgent: p.userAgent,
    locale: p.languages[0],
    timezoneId: p.timezone,
    viewport: { width: p.screen.width, height: p.screen.height },
    deviceScaleFactor: p.screen.dpr,
  };
  
  if (proxy) {
    options.proxy = { server: proxy.url };
    if (proxy.user && proxy.pass) {
      options.proxy.username = proxy.user;
      options.proxy.password = proxy.pass;
    }
  }

  const ctx = await browser.newContext(options);
  await ctx.addInitScript(spoofScript(p));
  return ctx;
}
