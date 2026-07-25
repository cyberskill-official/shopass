# extension (Shopass)

MV3 browser extension for Shopass.

## Setup

```bash
npm ci
```

Requires Node matching `.nvmrc` / `engines`. Then:

```bash
npm test
npm run typecheck
npm run build
```

Do not commit `node_modules`. Prefer `npm ci` over `npm install` so the lockfile is enforced.
