/**
 * When R28 sandboxes are live, flip this to true (or set NEXT_PUBLIC_CHECKOUT_LIVE=1)
 * so pricing CTAs go to /billing checkout instead of the waitlist modal.
 */
export function isCheckoutLive(): boolean {
  return process.env.NEXT_PUBLIC_CHECKOUT_LIVE === "1";
}
