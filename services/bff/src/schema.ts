import { readFileSync } from "node:fs";
import { join } from "node:path";
import { makeExecutableSchema } from "@graphql-tools/schema";
import type { GraphQLSchema } from "graphql";
import { wishlistResolvers } from "./resolvers/wishlist";
import { chartResolvers } from "./resolvers/chart";

// schema.graphql (copied next to the compiled output at build time) is the source
// of truth for the SDL; resolvers are attached here. Only Query is defined - no
// Mutation in this slice (DEC-WEB-25).
export function loadTypeDefs(): string {
  return readFileSync(join(__dirname, "schema.graphql"), "utf8");
}

const resolvers = {
  Query: {
    ...wishlistResolvers.Query,
    ...chartResolvers.Query,
  },
  WishlistItem: {
    ...wishlistResolvers.WishlistItem,
  },
};

export function makeSchema(): GraphQLSchema {
  return makeExecutableSchema({ typeDefs: loadTypeDefs(), resolvers });
}
