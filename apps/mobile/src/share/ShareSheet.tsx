import { buildUserShareLink } from "./shareLink";
import type { HttpClient } from "../api/httpClient";

export interface ShareSheetAction {
  userInitiated: true;
  url: string;
}

/**
 * Share-on-sale must be user-initiated (DEC-MOBILE-22).
 * Call only from an explicit "Chia sẻ" tap — never from a background timer.
 */
export async function onSharePressed(
  http: HttpClient,
  base: string,
  productId: number,
): Promise<ShareSheetAction> {
  const url = await buildUserShareLink(http, base, productId);
  return { userInitiated: true, url };
}
