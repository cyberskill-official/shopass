# Contributing (extension)

Thanks for helping improve Shopass.

## Setup

```bash
npm ci
npm test
npm run typecheck
npm run build
```

Node must match `engines` / `.nvmrc` (Node 24).

## Guidelines

1. Keep Manifest V3 constraints: no persistent background page, `chrome.alarms` ≥ 30s, no long-lived global state in the service worker.
2. Never send marketplace session tokens or passwords to Shopass servers — only product id / price / qty style payloads after consent.
3. Affiliate links activate only on explicit user click.
4. Prefer small PRs with tests for consent, DNR, and content-script parsing changes.
5. Do not commit `node_modules` or secrets.

## License

This package is MIT — see `LICENSE`.
