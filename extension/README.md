# Shopass browser extension

MV3 extension for [Shopass](https://shopass.cyberskill.world): cart/voucher helpers under consent, price capture, and user-initiated affiliate links.

**License:** MIT (`LICENSE`)  
**Privacy:** https://shopass.cyberskill.world/chinh-sach-bao-mat  
**Transparency:** https://shopass.cyberskill.world/minh-bach

## Setup

```bash
npm ci
npm test
npm run typecheck
npm run build          # production → dist/
npm run build:dev      # local API
```

Requires Node matching `engines` (≥24 &lt;25). Prefer `npm ci` so the lockfile is enforced. Do not commit `node_modules`.

## Load unpacked (Chrome / Edge / Cốc Cốc)

1. `npm run build` (or `build:dev` for local gateway).
2. Open `chrome://extensions` → Developer mode → **Load unpacked**.
3. Select the `dist/` directory produced by the build.

## Tests

```bash
npm test
```

DNR and consent suites live under `*.test.ts` next to the source.

## Store listing pack

See `store/` (R49): listing copy, data-disclosure worksheet, screenshots checklist.

## Security

See `SECURITY.md`. Report vulnerabilities to info@cyberskill.world.

## Contributing

See `CONTRIBUTING.md` and `REPRODUCIBLE-BUILD.md`.

## Public mirror (pending)

Going public as a standalone repo (`cyberskill/shopass-extension` or similar) requires founder approval (R36). Until then this tree lives in the Shopass monorepo under `extension/`.
