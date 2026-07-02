import { PageView, RawPrice, parseVNDInt } from "../types";
import { tiktokSelectors } from "./selectors";

const VND_UNIT = 1; // TikTok hiển thị VND nguyên; chỉnh nếu phát hiện micro-đơn-vị

function pickPriceFromState(state: any): RawPrice | null {
  try {
    // Trích xuất từ SIGI_STATE hoặc NEXT_DATA (giả định cấu trúc JSON phổ biến trên TikTok Shop)
    // Cấu trúc có thể là ItemModule hoặc ProductModule
    let price: number | null = null;
    let listPrice: number | null = null;
    let flashSale = false;

    if (state?.ItemModule) {
      const keys = Object.keys(state.ItemModule);
      const firstKey = keys[0];
      if (firstKey) {
        const item = state.ItemModule[firstKey];
        if (item?.price?.priceValue) price = Math.round(Number(item.price.priceValue) * VND_UNIT);
        if (item?.price?.originalPriceValue) listPrice = Math.round(Number(item.price.originalPriceValue) * VND_UNIT);
        if (item?.activityInfo?.isFlashSale) flashSale = true;
      }
    } else if (state?.props?.pageProps?.initialData?.productInfo) {
      const info = state.props.pageProps.initialData.productInfo;
      if (info?.price) price = Math.round(Number(info.price) * VND_UNIT);
      if (info?.originalPrice) listPrice = Math.round(Number(info.originalPrice) * VND_UNIT);
      if (info?.isFlashSale) flashSale = true;
    } else if (state?.ProductModule) {
      const p = state.ProductModule;
      if (p.price) price = Math.round(Number(p.price) * VND_UNIT);
      if (p.originalPrice) listPrice = Math.round(Number(p.originalPrice) * VND_UNIT);
      if (p.isFlashSale) flashSale = true;
    }

    if (price !== null) {
      return { price, listPrice, flashSale, source: "json" };
    }
  } catch (err) {
    // lỗi parse, fallback
  }
  return null;
}

// extractTikTok ưu tiên JSON nhúng, fallback DOM text; trả giá VND số nguyên.
export function extractTikTok(page: PageView): RawPrice {
  const stateStr = page.embeddedJSON(tiktokSelectors.embeddedState);
  if (stateStr) {
    try {
      const state = typeof stateStr === 'string' ? JSON.parse(stateStr) : stateStr;
      const p = pickPriceFromState(state); // đọc price/origin/flash từ state JSON
      if (p) return p;                      // source = json
    } catch (e) {
      // JSON parse error, ignore and fallback to DOM
    }
  }

  // fallback DOM text
  const priceText = page.text(tiktokSelectors.priceText);
  const listPriceText = page.text(tiktokSelectors.listPriceText);
  
  if (!priceText) {
    throw new Error("extract fail: Cannot find price in DOM");
  }

  return {
    price: parseVNDInt(priceText) || 0, // bỏ ký tự '₫', '.', số nguyên
    listPrice: parseVNDInt(listPriceText),
    flashSale: page.exists(tiktokSelectors.flashBadge),
    source: 'dom',
  };
}
