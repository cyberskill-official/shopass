export interface CartItem {
  productId: string;
  price: number;
  qty: number;
}

export interface VoucherItem {
  code: string;
  minSpend?: number;
  discountText?: string;
}

export interface ParseDomRequest {
  target: "offscreen";
  type: "PARSE_DOM";
  html: string;            // HTML THÔ đã có sẵn (content script lấy), KHÔNG phải URL
  platform: "shopee" | "tiktok" | "lazada";
}

export interface ParseDomResult {
  type: "PARSE_DOM_RESULT";
  items: Array<{ productId: string; price: number; qty: number }>;
}

export type Message =
  | { type: "CART_READ"; platform: "shopee"; items: CartItem[]; vouchers: VoucherItem[] }
  | { type: "PING" }
  | ParseDomRequest
  | ParseDomResult
  | TestCodesRequest;

// FR-EXT-006: Consent types (PDPL §5.5 compliance)
export type ConsentPurpose = "read_cart" | "read_voucher" | "sync_backend";

export interface ConsentRecord {
  policyVersion: string;     // e.g. "2026-06-27"
  decidedAt: number;         // epoch ms
  granted: ConsentPurpose[]; // mục đã đồng ý - tái lập được (DEC-EXT-31)
}

// FR-CART-005: Auto test codes
export interface CodeTestResult {
  code: string;
  discount: number;
}

export interface TestCodesRequest {
  type: "TEST_CODES";
  candidateCodes: string[];
  userInitiated: boolean;
}
