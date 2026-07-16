import { createServer, type IncomingMessage, type ServerResponse } from "node:http";
import {
  parse,
  validate,
  execute,
  specifiedRules,
  GraphQLError,
  type GraphQLSchema,
} from "graphql";
import { makeSchema } from "./schema";
import { buildContext, type ContextConfig } from "./context";
import { enforceLimits } from "./security/limits";

interface GraphQLBody {
  query?: string;
  variables?: Record<string, unknown> | null;
  operationName?: string | null;
}

function sendJSON(res: ServerResponse, status: number, body: unknown): void {
  const payload = JSON.stringify(body);
  res.writeHead(status, { "content-type": "application/json; charset=utf-8" });
  res.end(payload);
}

async function readBody(req: IncomingMessage): Promise<GraphQLBody> {
  const chunks: Buffer[] = [];
  for await (const c of req) chunks.push(c as Buffer);
  if (chunks.length === 0) return {};
  return JSON.parse(Buffer.concat(chunks).toString("utf8")) as GraphQLBody;
}

// handleGraphQL runs the read pipeline: parse -> depth/cost cap -> validate ->
// execute. Depth/cost caps reject before execution (TASK-WEB-005 §1 #4).
export async function handleGraphQL(
  schema: GraphQLSchema,
  cfg: ContextConfig,
  req: IncomingMessage,
  res: ServerResponse,
): Promise<void> {
  let body: GraphQLBody;
  try {
    body = await readBody(req);
  } catch {
    sendJSON(res, 400, { errors: [{ message: "invalid JSON body" }] });
    return;
  }
  if (!body.query) {
    sendJSON(res, 400, { errors: [{ message: "missing query" }] });
    return;
  }

  let document;
  try {
    document = parse(body.query);
  } catch (err) {
    const gql = err instanceof GraphQLError ? err : new GraphQLError(String(err));
    sendJSON(res, 400, { errors: [gql] });
    return;
  }

  try {
    enforceLimits(document); // depth + cost caps, before any resolver runs
  } catch (err) {
    const gql = err instanceof GraphQLError ? err : new GraphQLError(String(err));
    sendJSON(res, 400, { errors: [gql] });
    return;
  }

  const validationErrors = validate(schema, document, specifiedRules);
  if (validationErrors.length > 0) {
    sendJSON(res, 400, { errors: validationErrors });
    return;
  }

  const ctx = buildContext(req.headers, cfg);
  const result = await execute({
    schema,
    document,
    contextValue: ctx,
    variableValues: body.variables ?? undefined,
    operationName: body.operationName ?? undefined,
  });
  sendJSON(res, 200, result);
}

function env(k: string, def: string): string {
  const v = process.env[k];
  return v && v !== "" ? v : def;
}

export function startServer(): void {
  const schema = makeSchema();
  const cfg: ContextConfig = {
    trackBase: env("TRACK_BASE_URL", "http://track:8080"),
    dealBase: env("DEAL_BASE_URL", "http://deal:8080"),
  };
  const addr = env("BFF_ADDR", ":8085");
  const port = Number(addr.replace(/^.*:/, "")) || 8085;

  const server = createServer((req, res) => {
    if (req.method === "POST" && (req.url === "/graphql" || req.url?.startsWith("/graphql?"))) {
      handleGraphQL(schema, cfg, req, res).catch(() => {
        sendJSON(res, 500, { errors: [{ message: "internal server error" }] });
      });
      return;
    }
    if (req.method === "GET" && req.url === "/healthz") {
      sendJSON(res, 200, { ok: true });
      return;
    }
    sendJSON(res, 404, { errors: [{ message: "not found" }] });
  });

  server.listen(port, () => {
    // eslint-disable-next-line no-console
    console.log(`bff listening on :${port}`);
  });
}

if (require.main === module) {
  startServer();
}
