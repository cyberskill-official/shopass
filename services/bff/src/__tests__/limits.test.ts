import { test } from "node:test";
import assert from "node:assert/strict";
import { parse, GraphQLError } from "graphql";
import { enforceLimits, queryDepth, queryCost, MAX_COST } from "../security/limits";

test("a normal dashboard query passes the default caps", () => {
  const doc = parse(`{
    me { id displayName }
    wishlists { id name items { productId targetPrice chart { maturity daily { day closeP } } } }
  }`);
  assert.doesNotThrow(() => enforceLimits(doc));
});

test("queryDepth counts nesting; enforceLimits rejects when over the depth cap", () => {
  const doc = parse(`{ wishlists { items { chart { daily { day } } } } }`);
  // wishlists(1) items(2) chart(3) daily(4) day(5)
  assert.equal(queryDepth(doc), 5);
  // With a small explicit cap the same query is rejected before execution.
  try {
    enforceLimits(doc, { maxDepth: 3 });
    assert.fail("expected a depth error");
  } catch (err) {
    assert.ok(err instanceof GraphQLError);
    assert.equal((err as GraphQLError).extensions.code, "QUERY_TOO_DEEP");
  }
});

test("a wildly aliased query exceeds the cost cap and is rejected (DEC-WEB-24)", () => {
  // 400 aliased productChart fields, each pulling a daily list - far over MAX_COST.
  const fields = Array.from({ length: 400 }, (_, i) =>
    `c${i}: productChart(productId: "p${i}") { daily { day minP maxP closeP } }`,
  ).join("\n");
  const doc = parse(`{ ${fields} }`);
  assert.ok(queryCost(doc) > MAX_COST, `cost ${queryCost(doc)} should exceed ${MAX_COST}`);
  try {
    enforceLimits(doc);
    assert.fail("expected a cost error");
  } catch (err) {
    assert.ok(err instanceof GraphQLError);
    assert.equal((err as GraphQLError).extensions.code, "QUERY_TOO_COMPLEX");
  }
});
