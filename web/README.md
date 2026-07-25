# web (Shopass)

Next.js app for the Shopass product UI.

## Setup

Use the lockfile — do not improvise installs:

```bash
npm ci
```

Requires Node matching `.nvmrc` (and monorepo pin docs). Then:

```bash
npm run dev
npm test
npx tsc --noEmit
```

## Scripts

- `npm run dev` — local Next.js server
- `npm run build` / `npm start` — production build
- `npm test` — Jest
- `npm run lint` — ESLint
