import { PageView, RawPrice, parseVNDInt } from "../types";
import { lazadaSelectors } from "./selectors";

const VND_UNIT = 1;

function pickPriceFromModuleData(state: any): RawPrice | null {
  try {
    let price: number | null = null;
    let listPrice: number | null = null;
    let flashSale = false;

    // Lazada json state usually has "core" or "price" in it
    // Mocks based on general lazada structure
    if (state?.core?.price) {
      price = Math.round(Number(state.core.price) * VND_UNIT);
    } else if (state?.price?.current) {
      price = Math.round(Number(state.price.current) * VND_UNIT);
    }

    if (state?.core?.originPrice) {
      listPrice = Math.round(Number(state.core.originPrice) * VND_UNIT);
    } else if (state?.price?.original) {
      listPrice = Math.round(Number(state.price.original) * VND_UNIT);
    }

    if (state?.flashSale || state?.activity?.type === 'flash_sale') {
      flashSale = true;
    }

    if (price !== null) {
      return { price, listPrice, flashSale, source: "json" };
    }
  } catch (err) {
  }
  return null;
}

export function extractLazada(page: PageView): RawPrice {
  const stateStr = page.embeddedJSON(lazadaSelectors.embeddedState);
  if (stateStr) {
    try {
      let state = stateStr;
      if (typeof stateStr === 'string') {
        // Lazada often sets it like window.pageData = {...}
        const match = stateStr.match(/window\.(pageData|__moduleData__)\s*=\s*(\{.*?\});/);
        if (match && match[2]) {
          state = JSON.parse(match[2]);
        }
      }
      const p = pickPriceFromModuleData(state);
      if (p) return p;
    } catch (e) {
    }
  }

  const priceText = page.text(lazadaSelectors.priceText);
  const listPriceText = page.text(lazadaSelectors.listPriceText);
  
  if (!priceText) {
    throw new Error("extract fail: Cannot find price in DOM");
  }

  return {
    price: parseVNDInt(priceText) || 0,
    listPrice: parseVNDInt(listPriceText),
    flashSale: page.exists(lazadaSelectors.flashBadge),
    source: 'dom',
  };
}
