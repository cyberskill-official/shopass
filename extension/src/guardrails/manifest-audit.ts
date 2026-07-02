import { Manifest, Violation, viol } from './types';

export function auditManifest(m: Manifest): Violation[] {
  const out: Violation[] = [];
  if ((m.permissions ?? []).includes("cookies")) {
    out.push(viol("manifest", "cookies permission present", "least-privilege: no cookie access to shop hosts (§1 #4)"));
  }
  if ((m.permissions ?? []).includes("webRequestBlocking")) {
    out.push(viol("manifest", "webRequestBlocking present", "no blocking webRequest to rewrite redirects (§1 #4)"));
  }
  return out;
}
