import { test } from "node:test";
import assert from "node:assert/strict";
import { graphql } from "graphql";
import { makeSchema } from "../schema";
import { contextFromRest } from "../context";
import { FakeRest } from "./fakes";
import type { Wishlist } from "../types";

const schema = makeSchema();

function twoWishlists(): Wishlist[] {
  return [
    {
      id: "w1",
      name: "Tết",
      items: [
        { productId: "p1", targetPrice: 90000 },
        { productId: "p2", targetPrice: null },
      ],
    },
    {
      id: "w2",
      name: "Điện tử",
      items: [
        { productId: "p2", targetPrice: 500000 }, // duplicate of w1's p2
        { productId: "p3", targetPrice: null },
      ],
    },
  ];
}

test("anonymous request is rejected on every root field (§1 #7)", async () => {
  const rest = new FakeRest({ wishlists: twoWishlists() });
  const ctx = contextFromRest(null, rest); // no gateway identity

  for (const source of [
    "{ me { id } }",
    "{ wishlists { id } }",
    `{ productChart(productId: "p1") { maturity } }`,
  ]) {
    const res = await graphql({ schema, source, contextValue: ctx });
    assert.ok(res.errors && res.errors.length > 0, `expected error for: ${source}`);
    assert.equal(res.errors![0]!.message, "UNAUTHENTICATED");
  }
  assert.equal(rest.chartCalls, 0, "no backend calls for an anonymous request");
});

test("resolvers delegate to REST and forward the TASK-DEAL-003 chart feed shape", async () => {
  const rest = new FakeRest({ wishlists: twoWishlists() });
  const ctx = contextFromRest("u1", rest);
  const source = `{
    me { id displayName }
    wishlists { id name items { productId targetPrice } }
    productChart(productId: "p9") {
      maturity
      daily { day minP maxP closeP }
      annotations { median90 trailingMin verdict accumulating doubleDates }
    }
  }`;
  const res = await graphql({ schema, source, contextValue: ctx });
  assert.equal(res.errors, undefined);
  const data = res.data as any;
  assert.equal(data.me.id, "u1");
  assert.equal(data.me.displayName, "Chi");
  assert.equal(data.wishlists.length, 2);
  assert.equal(data.wishlists[0].items[0].productId, "p1");
  assert.equal(data.productChart.maturity, "mature");
  assert.equal(data.productChart.daily[0].closeP, 110000);
  assert.equal(data.productChart.annotations.verdict, "wait");
  assert.deepEqual(data.productChart.annotations.doubleDates, ["2026-06-01"]);
});

test("DataLoader batches + caches chart loads (anti-N+1, DEC-WEB-23)", async () => {
  const rest = new FakeRest({ wishlists: twoWishlists() });
  const ctx = contextFromRest("u1", rest);
  // 4 wishlist items (p1, p2, p2, p3) each request a chart, plus a standalone
  // productChart for p2 - naively 5 REST calls. With the loader it is one call
  // per distinct (productId, range): p1, p2, p3 -> 3.
  const source = `{
    wishlists { items { productId chart(range: "90d") { maturity } } }
    productChart(productId: "p2", range: "90d") { maturity }
  }`;
  const res = await graphql({ schema, source, contextValue: ctx });
  assert.equal(res.errors, undefined);
  assert.equal(rest.chartCalls, 3, `expected 3 distinct chart calls, got ${rest.chartCalls}`);
  assert.deepEqual([...rest.chartKeys].sort(), ["p1:90d", "p2:90d", "p3:90d"]);
});

test("an upstream 404/403 maps to a neutral error, no data leak (§1 #10)", async () => {
  const rest = new FakeRest({ wishlists: twoWishlists(), failChart: true });
  const ctx = contextFromRest("u1", rest);
  const res = await graphql({
    schema,
    source: `{ productChart(productId: "p1") { maturity } }`,
    contextValue: ctx,
  });
  assert.ok(res.errors && res.errors.length > 0);
  assert.equal(res.errors![0]!.message, "NOT_FOUND");
  // productChart is non-null (ChartData!), so a resolver error nullifies data.
  assert.equal(res.data, null);
});
