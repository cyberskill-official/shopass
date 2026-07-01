// extension/src/content/shared/test-codes.ts
import { randomDelayMs, sleep } from "./pacing";
import { userInitiatedApply, revert } from "./code-applier";
import { CodeTestResult } from "../../shared/types";

// testCodes bám đúng pseudo-code §3.5(4) (DEC-CART-24..29).
// CHỈ gọi từ user gesture (nút "thử mã"); `userInitiated` phải true.
export async function testCodes(
  candidateCodes: string[],     // từ voucher_catalog còn hiệu lực (DEC-CART-29)
  opts: { userInitiated: boolean; cancelled: () => boolean }
): Promise<CodeTestResult[]> {
  if (!opts.userInitiated) {
    throw new Error("testCodes chỉ chạy khi user chủ động bấm (DEC-CART-24)");
  }
  const results: CodeTestResult[] = [];
  for (const code of candidateCodes) {
    if (opts.cancelled()) break;            // user rời/hủy -> dừng (§1 #10)
    await sleep(randomDelayMs());           // nhịp người (DEC-CART-25)
    
    const apply = await userInitiatedApply(code);
    if (apply.valid) {
      results.push({ code, discount: apply.discount });
    }
    revert();                               // gỡ mã, KHÔNG chốt đơn (DEC-CART-26)
  }
  
  return results.sort((a, b) => b.discount - a.discount); // sortDesc (DEC-CART-27)
}
