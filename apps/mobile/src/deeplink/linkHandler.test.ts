import { parseDeepLink, routeFromDeepLink } from "./linkHandler";

describe("linkHandler", () => {
  it("parses product + ref", () => {
    expect(parseDeepLink("https://shopass.app/p?product_id=3&ref=R1")).toEqual({
      productId: 3,
      ref: "R1",
    });
  });

  it("drops malformed ref but keeps product", () => {
    const link = parseDeepLink("https://shopass.app/p?product_id=3&ref=!!!bad");
    expect(link.productId).toBe(3);
    expect(link.ref).toBeUndefined();
    expect(routeFromDeepLink(link)).toBe("Product");
  });

  it("routes home when product missing", () => {
    expect(routeFromDeepLink(parseDeepLink("https://shopass.app/p?ref=R1"))).toBe("Home");
  });
});
