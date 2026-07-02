export interface RawPrice {
  price: number;
  listPrice: number | null;
  flashSale: boolean;
  source: 'json' | 'dom';
}

export interface PageView {
  embeddedJSON(selector: string): any | null;
  text(selector: string): string | null;
  exists(selector: string): boolean;
}

export function parseVNDInt(str: string | null): number | null {
  if (!str) return null;
  const cleaned = str.replace(/[^\d]/g, '');
  if (!cleaned) return null;
  return parseInt(cleaned, 10);
}
