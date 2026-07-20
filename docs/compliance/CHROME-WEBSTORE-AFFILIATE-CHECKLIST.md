# Chrome Web Store Affiliate Policy Checklist (2025-03)

This checklist maps the requirements of the Chrome Web Store Affiliate Marketing policy (updated March 2025, enforced June 10, 2025) to our automated CI guardrails.

## Policy Requirements & Enforcement

| Policy Requirement | Enforcement Guardrail | Test Suite |
|---|---|---|
| **No Cookie Stuffing/Dropping** <br> Cannot silently set affiliate cookies on merchant domains without a click. | Static AST scan blocking `chrome.cookies.set` and `document.cookie` on shop hosts. | `extension/test/guardrails/no-cookie-stuffing.test.ts` |
| **No Pop-unders or Background Redirects** <br> Affiliate links cannot be opened invisibly or in the background. | Static AST scan blocking `window.open` with affiliate patterns, and preventing generic `chrome.tabs.update`. | `extension/test/guardrails/no-cookie-stuffing.test.ts` |
| **Must Require User Action** <br> Every affiliate redirection must be initiated by an explicit user gesture (e.g. click). | Runtime assertion checking `Event.isTrusted` before `window.open`. | `extension/test/guardrails/single-affiliate-path.test.ts` |
| **Least Privilege Manifest** <br> Extension cannot request permissions unnecessarily broad for affiliate tracking. | Manifest audit blocking `cookies` permission on shop hosts and `webRequestBlocking`. | `extension/test/guardrails/manifest-audit.test.ts` |
| **Clear Disclosure & Intent** <br> Affiliate functionality must be disclosed, and there can be no hidden back-doors. | Go backend asserting strictly ONE valid route for link generation and requiring `IncludesDisclosure=true`. | `services/affil/internal/affil/guardrail_assert_test.go` |

## Enforcement Mechanism
All of these guardrails are enforced as **CI Gates**. If any test fails, the build turns red and the pull request cannot be merged. This translates a policy promise into a verifiable technical invariant.
