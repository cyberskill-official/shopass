/**
 * dnr.ts — sanity-check wrapper for declarativeNetRequest rules.
 * MUST NOT use webRequest blocking (DEC-EXT-20).
 * Rules MUST be a minimal static set (DEC-EXT-21).
 *
 * R31: rules.json holds audited API allow rules (≤5) aligned with
 * shared/allowlist.ts. host_permissions remain the hard Chrome boundary for
 * outbound hosts; DNR documents the Shopass API allowlist for reviewers.
 * Do not add block/redirect rules aimed at marketplace domains (DEC-EXT-21).
 */

const MAX_ALLOWED_RULES = 5; // tối thiểu, đếm được, audit được

/**
 * Kiểm số lượng + phạm vi rule DNR lúc khởi động.
 * Trả true nếu rule count nằm trong ngưỡng tối thiểu.
 */
export async function validateDnrRules(): Promise<{
  valid: boolean;
  ruleCount: number;
}> {
  if (
    typeof chrome === "undefined" ||
    !chrome.declarativeNetRequest?.getDynamicRules
  ) {
    return { valid: true, ruleCount: 0 };
  }
  const rules = await chrome.declarativeNetRequest.getDynamicRules();
  return {
    valid: rules.length <= MAX_ALLOWED_RULES,
    ruleCount: rules.length,
  };
}

/**
 * Đếm static rules đã khai báo trong manifest.
 */
export async function getStaticRuleCount(): Promise<number> {
  if (
    typeof chrome === "undefined" ||
    !chrome.declarativeNetRequest?.getSessionRules
  ) {
    return 0;
  }
  const rules = await chrome.declarativeNetRequest.getSessionRules();
  return rules.length;
}

export { MAX_ALLOWED_RULES };
