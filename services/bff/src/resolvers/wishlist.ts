import { GraphQLError } from "graphql";
import type { GqlContext } from "../context";
import type { WishlistItem } from "../types";

function requireUser(ctx: GqlContext): void {
  // Every resolver requires a user context from the gateway (TASK-WEB-005 §1 #7):
  // no personal data for an anonymous request (one that never passed the gateway).
  if (!ctx.userId) {
    throw new GraphQLError("UNAUTHENTICATED", { extensions: { code: "UNAUTHENTICATED" } });
  }
}

export const wishlistResolvers = {
  Query: {
    me: (_: unknown, __: unknown, ctx: GqlContext) => {
      requireUser(ctx);
      return ctx.rest.getMe();
    },
    wishlists: (_: unknown, __: unknown, ctx: GqlContext) => {
      requireUser(ctx);
      // track-svc enforces ownership (DEC-WEB-22); the BFF does not re-implement it.
      return ctx.rest.listWishlists();
    },
  },
  WishlistItem: {
    // Loaded through the per-request DataLoader so a wishlist of N items makes one
    // batched round of chart calls, not N sequential ones (DEC-WEB-23).
    chart: (item: WishlistItem, args: { range?: string }, ctx: GqlContext) =>
      ctx.loaders.chart.load({ productId: item.productId, range: args.range ?? "90d" }),
  },
};
