# Shopass VPS checklist (closed beta)

Use after each prod image roll or host reboot. Prod SHA today is tracked in
`docs/tasks/improvement/LEDGER.md` — do not invent deploy authority.

## Disk / capacity

```bash
df -h /
docker system df
# Before rebuilding web locally: free ≥ 3–4 GiB. Prefer GHCR pull (R15) over build-on-box.
# Safe prune (dangling/unused only — never prune named volumes):
docker image prune -f
```

Alert: root filesystem should stay under ~80%. Backup dumps live in
`/var/backups/shopass` (mode `0700`).

## Edge + compose

```bash
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f /srv/shopass/deploy/docker-compose.production.yml \
  -f /srv/shopass/deploy/docker-compose.cyberos-shared.yml \
  ps
# CyberOS Caddy on shopass-edge:
systemctl is-enabled shopass-edge-attach.service
curl -fsS -o /dev/null -w '%{http_code}\n' https://shopass.cyberskill.world/
```

## Timers (R12 + R17)

```bash
systemctl list-timers 'shopass-*'
sudo systemctl status shopass-backup.timer shopass-scrape.timer shopass-forecast.timer --no-pager
# One-shot backup (local; upload needs BACKUP_S3_* — see RESTORE-RUNBOOK):
sudo /srv/shopass/deploy/scripts/backup-pg.sh
# Dry-run upload tooling without writing objects:
sudo /srv/shopass/deploy/scripts/backup-upload-dry-run.sh
```

## Monitoring

- Prometheus rules: `ShopassBackupStale` (26h) in `deploy/prometheus/rules/shopass.yml`
- Pushgateway heartbeat from backup/scrape/forecast scripts
- Observability compose is optional overlay — see `deploy/README.md`

## Banned on prod

- `make smoke` / `make seed` / `ALLOW_SEED=1`
- Enabling Google OAuth (`ENABLE_GOOGLE_OAUTH`)
- Pasting object-storage or SMTP secrets into git/PRs
