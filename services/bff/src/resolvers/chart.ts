import { GraphQLError } from "graphql";
import type { GqlContext } from "../context";

export const chartResolvers = {
  Query: {
    productChart: (
      _: unknown,
      args: { productId: string; range?: string },
      ctx: GqlContext,
    ) => {
      if (!ctx.userId) {
        throw new GraphQLError("UNAUTHENTICATED", { extensions: { code: "UNAUTHENTICATED" } });
      }
      // Through the same DataLoader as WishlistItem.chart, so a dashboard query
      // that asks for a wishlist and a standalone chart of the same product hits
      // the backend once (FR-WEB-005 §1 #3). The BFF forwards the FR-DEAL-003 feed
      // shape unchanged - it does not compute verdict/median (§1 #9).
      return ctx.loaders.chart.load({ productId: args.productId, range: args.range ?? "90d" });
    },
  },
};
