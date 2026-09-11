# Shopass restore runbook (R12)

## Targets

| Metric | Wave-1 (nightly dump) | Follow-up (WAL / PITR) |
|--------|----------------------|-------------------------|
| RPO | **24h** (nightly dump) | ~5 min with wal-g/pgBackRest |
| RTO | **≤ 1h** (restore dump into new DB + cutover) | ≤ 1h |

PITR (wal-g or pgBackRest) is intentionally deferred; track as R12.1 when object
storage and WAL archiving are funded.

## What gets backed up

- Nightly logical dump: `pg_dump | gzip` via `deploy/scripts/backup-pg.sh`
- Filename: `shopass_YYYY-MM-DD_<schema-git-sha>.sql.gz`
- Local dir: `/var/backups/shopass` (mode `0700`)
- Optional off-host: `BACKUP_S3_URI` (rclone remote or `s3://bucket/prefix`)
- Retention: 30 days local (+ remote delete when using rclone)
- Heartbeat: `shopass_job_last_success_unixtime{job_name="backup"}` → Pushgateway
  (alert `ShopassBackupStale` if > 26h)

## Install (host)

```bash
sudo install -d -m 0700 /var/backups/shopass
sudo install -m 0755 deploy/scripts/backup-pg.sh /srv/shopass/deploy/scripts/
sudo install -m 0755 deploy/scripts/restore-drill.sh /srv/shopass/deploy/scripts/
sudo install -m 0644 deploy/systemd/shopass-backup.service /etc/systemd/system/
sudo install -m 0644 deploy/systemd/shopass-backup.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now shopass-backup.timer
systemctl list-timers 'shopass-backup*'
```

Add to `/etc/shopass/runtime.env` when Stephen provisions object storage:

```dotenv
BACKUP_LOCAL_DIR=/var/backups/shopass
BACKUP_RETENTION_DAYS=30
# rclone example (preferred on VPS):
BACKUP_S3_URI=b2:shopass-backups/pg
BACKUP_UPLOAD_TOOL=rclone
# or AWS-compatible:
# BACKUP_S3_URI=s3://shopass-backups/pg
# BACKUP_UPLOAD_TOOL=aws
# AWS_ACCESS_KEY_ID=...
# AWS_SECRET_ACCESS_KEY=...
# AWS_DEFAULT_REGION=...
```

Install `rclone` (or AWS CLI) and configure the remote before enabling upload.
Without `BACKUP_S3_URI`, dumps stay local-only — **not** disaster-ready.

## Manual one-shot dump (no timer)

```bash
sudo /srv/shopass/deploy/scripts/backup-pg.sh
# or equivalent:
sudo docker compose --env-file /etc/shopass/runtime.env \
  -f /srv/shopass/deploy/docker-compose.production.yml exec -T db \
  pg_dump -U shopass shopass | gzip > /var/backups/shopass/shopass_manual_$(date -u +%F).sql.gz
```

## Restore drill (scratch — never prod)

```bash
LATEST=$(ls -1t /var/backups/shopass/shopass_*.sql.gz | head -1)
sudo DRILL_KEEP=0 /srv/shopass/deploy/scripts/restore-drill.sh "$LATEST"
```

Record in `docs/tasks/improvement/LEDGER.md`: dump bytes, elapsed seconds,
row counts (`app_user`, `tracked_product`), and whether upload to object storage
succeeded.

## Production restore (incident — operator only)

1. Freeze writes / take stack offline if corruption is confirmed.
2. Provision a new Timescale volume (or wipe only after verified backup).
3. Start empty `db`, create extension, restore:

```bash
gunzip -c /path/to/shopass_YYYY-MM-DD_SHA.sql.gz | \
  docker compose --env-file /etc/shopass/runtime.env \
    -f deploy/docker-compose.production.yml exec -T db \
    psql -U shopass -d shopass
```

4. Run migrations only if the dump predates current schema (forward-only).
5. Smoke: gateway `/healthz`, web `/api/healthz`, one authenticated chart read.
6. Do **not** run `make smoke` / seed against production.

## Stephen ask (off-host)

Until bucket + keys exist, R12 stays `needs_stephen`:

- Provider: Backblaze B2, Wasabi, Vultr Object Storage, or S3
- Bucket name + region/endpoint
- Access key + secret (or rclone config) installed on the VPS under
  `/etc/shopass/` (mode `0600`) or rclone remote for root
- Confirm 30-day retention is acceptable
