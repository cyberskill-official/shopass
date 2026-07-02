import { test } from "node:test";
import assert from "node:assert/strict";
import { GraphQLError } from "graphql";
import { HttpRestClient } from "../rest";

function fakeFetch(
  opts: { status?: number; ok?: boolean; json?: unknown; throwNetwork?: boolean },
  captured?: { headers?: Record<string, string>; url?: string },
): typeof fetch {
  return (async (url: string, init?: { headers?: Record<string, string> }) => {
    if (captured) {
      captured.url = url;
      captured.headers = init?.headers ?? {};
    }
    if (opts.throwNetwork) throw new Error("ECONNREFUSED");
    return {
      ok: opts.ok ?? true,
      status: opts.status ?? 200,
      json: async () => opts.json ?? {},
    } as unknown as Response;
  }) as unknown as typeof fetch;
}

async function codeOf(p: Promise<unknown>): Promise<unknown> {
  try {
    await p;
    return "NO_ERROR";
  } catch (err) {
    assert.ok(err instanceof GraphQLError, "must be a GraphQLError");
    return (err as GraphQLError).extensions.code;
  }
}

test("403 and 404 from a REST service map to a neutral NOT_FOUND", async () => {
  for (const status of [403, 404]) {
    const c = new HttpRestClient("u1", "r1", "http://track", "http://deal", fakeFetch({ ok: false, status }));
    assert.equal(await codeOf(c.listWishlists()), "NOT_FOUND");
    assert.equal(await codeOf(c.getChart("p1", "90d")), "NOT_FOUND");
  }
});

test("a 5xx or network failure maps to a generic UPSTREAM_ERROR, not swallowed", async () => {
  const c500 = new HttpRestClient("u1", "r1", "http://track", "http://deal", fakeFetch({ ok: false, status: 502 }));
  assert.equal(await codeOf(c500.getChart("p1", "90d")), "UPSTREAM_ERROR");

  const cNet = new HttpRestClient("u1", "r1", "http://track", "http://deal", fakeFetch({ throwNetwork: true }));
  assert.equal(await codeOf(cNet.listWishlists()), "UPSTREAM_ERROR");
});

test("forwards x-user-id and x-request-id downstream (§1 #1, #8)", async () => {
  const captured: { headers?: Record<string, string>; url?: string } = {};
  const c = new HttpRestClient(
    "user-42",
    "req-99",
    "http://track",
    "http://deal",
    fakeFetch({ ok: true, json: [] }, captured),
  );
  await c.listWishlists();
  assert.equal(captured.headers?.["x-user-id"], "user-42");
  assert.equal(captured.headers?.["x-request-id"], "req-99");
  assert.equal(captured.url, "http://track/v1/wishlists");
});
