import { GraphQLError, type DocumentNode, type SelectionSetNode, type FragmentDefinitionNode } from "graphql";

// Query depth and cost caps (TASK-WEB-005 §1 #4 / DEC-WEB-24). GraphQL lets a
// client shape its own query, including deeply nested or wildly aliased queries
// that exhaust resources - the classic GraphQL DoS surface. These caps reject
// such a query before any resolver runs.
export const MAX_DEPTH = 8;
export const MAX_COST = 1000;

// Fields that return a list; each carries extra weight because a list fans out.
// The cost is additive (weight + children), not multiplicative, so a normal
// two-level query (wishlist -> items -> chart) stays well under the cap while a
// wildly aliased query still blows past it. Deep nesting is caught by the depth
// cap, so cost does not also need to punish depth.
const LIST_FIELDS = new Set(["wishlists", "items", "daily", "doubleDates"]);
const LIST_WEIGHT = 10;

type FragMap = Record<string, FragmentDefinitionNode>;

function fragmentsOf(doc: DocumentNode): FragMap {
  const m: FragMap = {};
  for (const def of doc.definitions) {
    if (def.kind === "FragmentDefinition") m[def.name.value] = def;
  }
  return m;
}

function selectionSetDepth(sel: SelectionSetNode, frags: FragMap): number {
  let max = 0;
  for (const s of sel.selections) {
    if (s.kind === "Field") {
      const d = s.selectionSet ? 1 + selectionSetDepth(s.selectionSet, frags) : 1;
      if (d > max) max = d;
    } else if (s.kind === "InlineFragment" && s.selectionSet) {
      const d = selectionSetDepth(s.selectionSet, frags);
      if (d > max) max = d;
    } else if (s.kind === "FragmentSpread") {
      const fd = frags[s.name.value];
      if (fd) {
        const d = selectionSetDepth(fd.selectionSet, frags);
        if (d > max) max = d;
      }
    }
  }
  return max;
}

function selectionSetCost(sel: SelectionSetNode, frags: FragMap): number {
  let sum = 0;
  for (const s of sel.selections) {
    if (s.kind === "Field") {
      const child = s.selectionSet ? selectionSetCost(s.selectionSet, frags) : 0;
      const weight = LIST_FIELDS.has(s.name.value) ? LIST_WEIGHT : 1;
      sum += weight + child;
    } else if (s.kind === "InlineFragment" && s.selectionSet) {
      sum += selectionSetCost(s.selectionSet, frags);
    } else if (s.kind === "FragmentSpread") {
      const fd = frags[s.name.value];
      if (fd) sum += selectionSetCost(fd.selectionSet, frags);
    }
  }
  return sum;
}

// queryDepth returns the deepest field nesting across the document's operations.
export function queryDepth(doc: DocumentNode): number {
  const frags = fragmentsOf(doc);
  let max = 0;
  for (const def of doc.definitions) {
    if (def.kind === "OperationDefinition") {
      const d = selectionSetDepth(def.selectionSet, frags);
      if (d > max) max = d;
    }
  }
  return max;
}

// queryCost returns a weighted field count, with list fields inflating their subtree.
export function queryCost(doc: DocumentNode): number {
  const frags = fragmentsOf(doc);
  let sum = 0;
  for (const def of doc.definitions) {
    if (def.kind === "OperationDefinition") {
      sum += selectionSetCost(def.selectionSet, frags);
    }
  }
  return sum;
}

export interface Limits {
  maxDepth?: number;
  maxCost?: number;
}

// enforceLimits throws a GraphQLError (rejecting the query before execution) when
// it is too deep or too complex. Limits default to MAX_DEPTH / MAX_COST; callers
// (tests) may pass explicit limits.
export function enforceLimits(doc: DocumentNode, limits: Limits = {}): void {
  const maxDepth = limits.maxDepth ?? MAX_DEPTH;
  const maxCost = limits.maxCost ?? MAX_COST;
  const depth = queryDepth(doc);
  if (depth > maxDepth) {
    throw new GraphQLError(`query quá sâu (>${maxDepth})`, {
      extensions: { code: "QUERY_TOO_DEEP", depth },
    });
  }
  const cost = queryCost(doc);
  if (cost > maxCost) {
    throw new GraphQLError(`query quá phức tạp (>${maxCost})`, {
      extensions: { code: "QUERY_TOO_COMPLEX", cost },
    });
  }
}
