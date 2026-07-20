import { trustedSiteURL } from "./site-url";

const MAX_CAPTURE_PRICE_VND = 1_000_000_000_000;
const MAX_SHOPEE_URL_LENGTH = 4096;
const MAX_CAPTURE_PRICE_INPUT_LENGTH = 64;
const SHOPEE_PRODUCT_PATH = /(?:^|[-/])i\.\d+\.\d+(?:$|\/)/;

/**
 * Converts a price a person sees in their browser (for example
 * "6.490.000đ") into the integer we store.  We deliberately accept no
 * decimals: Shopass records prices in Vietnamese đồng.
 */
export function normalizeCapturedPrice(value: string | null | undefined): number | null {
  if (!value || value.length > MAX_CAPTURE_PRICE_INPUT_LENGTH) return null;

  const digits = (value ?? "").replace(/[^\d]/g, "");
  if (!digits || digits.length > 13) return null;

  const price = Number(digits);
  if (!Number.isSafeInteger(price) || price <= 0 || price > MAX_CAPTURE_PRICE_VND) {
    return null;
  }
  return price;
}

/**
 * Keep the browser-assisted route as narrow as the server-side parser.  The
 * backend still repeats this validation before it creates a tracked product.
 */
export function canonicalShopeeProductURL(value: string | null | undefined): string | null {
  if (!value || value.length > MAX_SHOPEE_URL_LENGTH) return null;

  try {
    const parsed = new URL(value);
    const hostname = parsed.hostname.toLowerCase();
    if (
      parsed.protocol !== "https:" ||
      (hostname !== "shopee.vn" && hostname !== "www.shopee.vn") ||
      parsed.username ||
      parsed.password ||
      parsed.port ||
      !SHOPEE_PRODUCT_PATH.test(parsed.pathname)
    ) {
      return null;
    }

    // Product IDs live in the path. Dropping tracking parameters and hashes
    // prevents a bookmarklet from carrying referral or session-like data into
    // Shopass while retaining the exact product identity we need.
    // Do not preserve a user-controlled origin, credentials, port, query, or
    // hash. The product identity lives only in the validated pathname.
    return new URL(parsed.pathname, "https://shopee.vn").toString();
  } catch {
    return null;
  }
}

/**
 * Creates a user-initiated bookmarklet. It performs no network request and
 * reads neither cookies nor account data. It only transfers a canonical
 * Shopee product URL plus a candidate visible price to Shopass, where the
 * signed-in user must explicitly confirm the write.
 */
export function buildShopeeCaptureBookmarklet(captureOrigin: string): string {
  const captureURL = new URL("/capture", trustedSiteURL(captureOrigin)).toString();
  const originLiteral = JSON.stringify(captureURL);

  const script = `(()=>{const h=location.hostname.toLowerCase();if((h!=="shopee.vn"&&h!=="www.shopee.vn")||location.protocol!=="https:"){alert("Mở một trang sản phẩm Shopee Việt Nam bằng HTTPS trước khi dùng nút Shopass.");return}const selectors=["meta[itemprop='price']","meta[property='product:price:amount']","[data-sqe='price']","[class*='product-price']","[class*='price']"];const read=n=>n.getAttribute("content")||n.textContent||"";const priceOf=v=>{const m=String(v).match(/\\d{1,3}(?:[.,\\s]\\d{3})+|\\d{4,12}/g)||[];for(const x of m){const n=Number(x.replace(/[^\\d]/g,""));if(Number.isSafeInteger(n)&&n>0&&n<=1000000000000)return String(n)}return""};let price="";for(const s of selectors){for(const n of Array.from(document.querySelectorAll(s)).slice(0,8)){price=priceOf(read(n));if(price)break}if(price)break}const productURL="https://shopee.vn"+location.pathname;if(productURL.length>4096||!/(?:^|[-/])i\\.\\d+\\.\\d+(?:$|\\/)/.test(location.pathname)){alert("Không nhận diện được liên kết sản phẩm Shopee này.");return}const target=new URL(${originLiteral});target.searchParams.set("url",productURL);if(price)target.searchParams.set("price",price);const a=document.createElement("a");a.href=target.toString();a.rel="noreferrer noopener";a.referrerPolicy="no-referrer";a.style.display="none";document.documentElement.appendChild(a);a.click();a.remove()})()`;

  return `javascript:${script}`;
}
