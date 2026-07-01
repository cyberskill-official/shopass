// extension/src/content/shared/code-applier.ts

export interface ApplyOutcome {
  valid: boolean;
  discount: number; // VND int
}

// In a real implementation, these would interact with the specific platform's DOM.
// We provide placeholders that represent the required action.

function enterCodeIntoCartField(code: string): void {
  // placeholder
}

function clickApplyButton(): void {
  // placeholder
}

function readDiscountFromCart(): number {
  // placeholder
  return 0;
}

function removeAppliedCodeFromCart(): void {
  // placeholder
}

export async function userInitiatedApply(code: string): Promise<ApplyOutcome> {
  enterCodeIntoCartField(code);   // gõ mã vào ô voucher của giỏ
  clickApplyButton();             // bấm "áp dụng" trên UI sàn
  
  // In reality, you'd wait for DOM changes here before reading
  const discount = readDiscountFromCart(); // đọc mức giảm hiển thị
  
  return { valid: discount > 0, discount };
  // LƯU Ý: KHÔNG có bước thực hiện giao dịch (DEC-CART-26/28)
}

export function revert(): void {
  removeAppliedCodeFromCart();    // gỡ mã ra khỏi giỏ (DEC-CART-26)
}
