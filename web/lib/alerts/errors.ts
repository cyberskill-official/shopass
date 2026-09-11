/** Thrown when creating a Premium-gated alert rule on Free (HTTP 402). */
export class PremiumRequiredError extends Error {
  readonly status = 402 as const;

  constructor(message = "Tính năng này cần Premium") {
    super(message);
    this.name = "PremiumRequiredError";
  }
}

export function isPremiumRequiredError(error: unknown): error is PremiumRequiredError {
  return error instanceof PremiumRequiredError;
}
