export function openAffiliate(ev: Event | undefined | null, deepLink: string): void {
  // @ts-ignore
  if (!ev || !ev.isTrusted) {
    throw new Error("affiliate navigation must originate from a trusted user gesture"); // §1 #3
  }
  // mở trong tab user bấm; KHÔNG set cookie, KHÔNG nền (cookie do trang sàn set)
  window.open(deepLink, "_blank", "noopener");
}
