import { test } from "node:test";
import assert from "node:assert/strict";
import { createServer, type IncomingMessage, type ServerResponse } from "node:http";
import { handleGraphQL, MAX_BODY_BYTES } from "../server";
import { makeSchema } from "../schema";
import type { ContextConfig } from "../context";

const cfg: ContextConfig = {
  trackBase: "http://track.test",
  dealBase: "http://deal.test",
};

test("oversized GraphQL body is rejected with 413", async () => {
  const schema = makeSchema();
  const server = createServer((req: IncomingMessage, res: ServerResponse) => {
    handleGraphQL(schema, cfg, req, res).catch(() => {
      res.writeHead(500);
      res.end();
    });
  });
  await new Promise<void>((resolve) => server.listen(0, resolve));
  const addr = server.address();
  assert.ok(addr && typeof addr === "object");
  const port = addr.port;

  const oversized = "x".repeat(MAX_BODY_BYTES + 1);
  const body = JSON.stringify({ query: `{ me { id } }`, pad: oversized });
  assert.ok(Buffer.byteLength(body) > MAX_BODY_BYTES);

  const res = await fetch(`http://127.0.0.1:${port}/graphql`, {
    method: "POST",
    headers: {
      "content-type": "application/json",
      "content-length": String(Buffer.byteLength(body)),
    },
    body,
  });
  assert.equal(res.status, 413);
  const json = (await res.json()) as { errors: Array<{ message: string }> };
  assert.match(json.errors[0]!.message, /too large/i);

  await new Promise<void>((resolve, reject) =>
    server.close((err) => (err ? reject(err) : resolve())),
  );
});
