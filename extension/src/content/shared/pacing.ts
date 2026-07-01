// extension/src/content/shared/pacing.ts
// sleep random 2.5s..5s giữa các lần thử (DEC-CART-25, §3.5(4)).
export function randomDelayMs(): number {
  return 2500 + Math.floor(Math.random() * 2500); // [2500, 5000)
}

export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
