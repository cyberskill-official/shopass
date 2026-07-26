/** Persist pending ref until signup; never overwrite an existing referrer. */

export class PendingReferral {
  private pending: string | null = null;
  private locked = false;

  setIfEmpty(ref: string): void {
    if (this.locked || this.pending) return;
    if (!ref || ref.trim() === "") return;
    this.pending = ref.trim();
  }

  /** Client-side self-referral block; backend remains final gate. */
  consume(selfCode: string): string | null {
    if (!this.pending) return null;
    if (this.pending === selfCode) {
      this.pending = null;
      return null;
    }
    const out = this.pending;
    this.pending = null;
    this.locked = true;
    return out;
  }
}
