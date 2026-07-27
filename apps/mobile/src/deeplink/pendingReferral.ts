/** Pending referral from deep-link until signup attribution (DEC-MOBILE-23/24). */

export class PendingReferral {
  private ref: string | null = null;

  setIfEmpty(ref: string): void {
    if (!this.ref) this.ref = ref;
  }

  peek(): string | null {
    return this.ref;
  }

  clear(): void {
    this.ref = null;
  }

  /**
   * Consume pending ref unless it equals the signed-in user's own code (self-referral).
   */
  consume(myReferralCode?: string): string | null {
    if (!this.ref) return null;
    if (myReferralCode && this.ref === myReferralCode) {
      this.ref = null;
      return null;
    }
    const out = this.ref;
    this.ref = null;
    return out;
  }
}
