# Reproducible build notes

Compare a Chrome Web Store / sideload zip against a local source build.

## Build

```bash
npm ci
npm run build          # SHOPASS_ENV=production → dist/
```

Output directory: `extension/dist/` (see `build.mjs`).

## Package zip (local)

```bash
cd dist
zip -r ../shopass-extension-src.zip .
cd ..
shasum -a 256 shopass-extension-src.zip
```

## Diff against a store package

1. Download the published CRX/zip from the store (or your submission artifact).
2. Unpack both archives to temp dirs.
3. Diff recursively (expect differences only in store-injected metadata if any):

```bash
diff -ru unpacked-store/ unpacked-src/ | head -200
```

4. Confirm `manifest.json` `version`, `icons`, `default_locale`, and host permissions match the listing worksheet in `store/`.

CI runs `npm test` + `tsc` on every PR; public mirror CI (R36 publish) should run the same suite.
