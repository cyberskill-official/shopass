import {
  displayOptimizeResult,
  type CartLine,
  type OptimizeResult,
} from "./optimizerClient";

export interface CheckoutAssistantModel {
  userInitiated: true;
  autoApply: false;
  summary: string;
  disclosure: string;
}

/** View model for user-initiated checkout assist — display only. */
export function buildCheckoutAssistant(
  result: OptimizeResult,
  affiliateDisclosure = true,
): CheckoutAssistantModel {
  return {
    userInitiated: true,
    autoApply: false,
    summary: displayOptimizeResult(result),
    disclosure: affiliateDisclosure
      ? "Có thể chứa liên kết affiliate. Bạn tự áp mã trong app sàn."
      : "",
  };
}

export type { CartLine };
