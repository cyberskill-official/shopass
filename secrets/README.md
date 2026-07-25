# secrets

Go library for Vault and AWS Secrets Manager providers (`Provider`, caching, masking).

## Adoption status

**No service imports this module yet.** Deploy still injects credentials via
environment files / Compose secrets (see `deploy/secrets/README.md`). Path
layout and least-privilege policy live in [`PATH-LAYOUT.md`](PATH-LAYOUT.md).

Wire a service to `shopass/secrets` only when it gains `_FILE` / Vault loading
at startup; until then treat this package as library-ready but unused in the
runtime topology.
