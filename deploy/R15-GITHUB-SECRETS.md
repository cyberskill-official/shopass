# R15 — GitHub secrets for GHCR publish + SSH deploy

Scaffolding ships in-repo. **Do not paste secrets into the PR.** Stephen (or org
admin) sets these in GitHub → Settings → Secrets and variables → Actions
(and optionally Environments → `production`).

Until secrets exist, workflows stay fail-closed: publish may skip push; deploy
job exits 0 with a clear “secrets missing” message and does **not** SSH.

## Required for image publish (`publish-ghcr.yml`)

| Secret / permission | Purpose |
|---------------------|---------|
| `GITHUB_TOKEN` (automatic) | Push to `ghcr.io` when `packages: write` is granted on the workflow |
| Org package visibility | Ensure `cyberskill-official/shopass-*` packages are create-able by Actions |

Images (tag = full git SHA + `latest` on `main`):

- `ghcr.io/cyberskill-official/shopass-gateway`
- `ghcr.io/cyberskill-official/shopass-authsvc`
- `ghcr.io/cyberskill-official/shopass-pricesvc`
- `ghcr.io/cyberskill-official/shopass-dealsvc`
- `ghcr.io/cyberskill-official/shopass-notifsvc`
- `ghcr.io/cyberskill-official/shopass-tracksvc`
- `ghcr.io/cyberskill-official/shopass-billsvc`
- `ghcr.io/cyberskill-official/shopass-complysvc`
- `ghcr.io/cyberskill-official/shopass-scrapesvc`
- `ghcr.io/cyberskill-official/shopass-web`
- `ghcr.io/cyberskill-official/shopass-mlforecast`
- `ghcr.io/cyberskill-official/shopass-playwright-farm` (node farm)

## Required for SSH deploy (`deploy.yml`, environment `production`)

| Secret | Example / notes |
|--------|-----------------|
| `SHOPASS_DEPLOY_HOST` | VPS hostname or IP (e.g. `shopass.cyberskill.world` or `203.x.x.x`) |
| `SHOPASS_DEPLOY_USER` | SSH user with docker compose rights (often `root` or `deploy`) |
| `SHOPASS_DEPLOY_SSH_KEY` | Private key PEM (ed25519 preferred). Public half in `~/.ssh/authorized_keys` on VPS |
| `SHOPASS_DEPLOY_PATH` | Repo checkout on host, default `/srv/shopass` |
| `SHOPASS_DEPLOY_COMPOSE` | Optional override; default uses production + ghcr overlay |

Optional:

| Secret | Purpose |
|--------|---------|
| `SHOPASS_DEPLOY_SSH_PORT` | Non-22 SSH |
| `SHOPASS_DEPLOY_KNOWN_HOSTS` | `ssh-keyscan` output; if unset, workflow uses `StrictHostKeyChecking=accept-new` once |

## VPS one-time prep (Stephen)

1. Checkout `/srv/shopass` (or chosen path); keep `git` remote readable.
2. Root-owned `/etc/shopass/runtime.env` mode `0600` from `deploy/.env.production.example`.
3. Install timers: see `deploy/systemd/README.md` + `deploy/VPS-CHECKLIST.md`.
4. `docker login ghcr.io` as a robot user **or** rely on public pulls if packages are public.
5. First manual roll (before enabling gated deploy):

```bash
cd /srv/shopass
git fetch && git checkout <sha>
export SHOPASS_IMAGE_TAG=<sha>
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f deploy/docker-compose.production.yml \
  -f deploy/docker-compose.ghcr.yml \
  pull
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f deploy/docker-compose.production.yml \
  -f deploy/docker-compose.ghcr.yml \
  up -d --remove-orphans
```

6. Smoke: `curl -fsS https://$APP_DOMAIN/api/healthz` and gateway `/healthz` via same-origin `/v1/` as appropriate. **Do not** run `make smoke` on prod.

## Status

R15 remains `needs_stephen` until secrets + first successful gated deploy are
HITL-accepted. R16 stays blocked on R15.
