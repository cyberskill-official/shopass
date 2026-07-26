# NOTICE — Shopass licensing split

**Product brand:** Shopass  
**Legal entity:** CyberSkill Software Solutions Consultancy and Development JSC  
**Public origin:** https://shopass.cyberskill.world  

## Split

| Path | License | File |
|------|---------|------|
| Repository root product code (`services/`, `web/`, `db/`, `deploy/`, `apps/`, `bff/`, etc.) | Proprietary | [`LICENSE`](LICENSE) |
| Browser extension (`extension/`) | MIT | [`extension/LICENSE`](extension/LICENSE) |

Operator decision (2026-07-26): proprietary core + MIT extension; brand name Shopass.

## Third-party dependencies

Go modules declare licenses via their upstream `LICENSE` / `NOTICE` files (see each `go.mod`).  
JavaScript dependencies under `web/`, `extension/`, and `services/bff/` ship their own licenses inside `node_modules/` after `npm ci`.

To regenerate a summary locally (not checked in — regenerates with dependency bumps):

```bash
# JS (from each package root)
npx --yes license-checker --summary

# Go (example)
go install github.com/google/go-licenses@latest
go-licenses report ./services/gateway/...
```

Chrome Web Store listing intent: yes (Wave 1), using this MIT-licensed extension tree.
