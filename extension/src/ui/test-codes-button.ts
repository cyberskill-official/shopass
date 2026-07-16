// extension/src/ui/test-codes-button.ts
import { testCodes } from "../content/shared/test-codes";
import { CodeTestResult } from "../shared/types";

// In a real scenario, this would fetch from the background which fetches from the catalog
async function getCandidateCodesFromCatalog(): Promise<string[]> {
  return []; // placeholder
}

function renderSuggestions(results: CodeTestResult[]): void {
  // placeholder
}

export function setupTestCodesButton(button: HTMLButtonElement): void {
  let userLeft = false;
  
  // Cleanup/cancellation logic would be more robust in reality
  window.addEventListener("beforeunload", () => {
    userLeft = true;
  });

  // Nút do người dùng bấm; KHÔNG tự gọi testCodes (DEC-CART-24).
  button.addEventListener("click", async () => {
    const codes = await getCandidateCodesFromCatalog(); // TASK-CART-001 còn hiệu lực
    const results = await testCodes(codes, { userInitiated: true, cancelled: () => userLeft });
    renderSuggestions(results); // gợi ý, user tự áp mã trên sàn
  });
}
