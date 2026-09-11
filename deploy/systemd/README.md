# Host scheduler units (R17)

The production Compose file keeps `scrapesvc` and `mlforecast` as one-shot
jobs. These systemd timers run them from the host instead of granting a
container the Docker socket.

After each successful run, `ExecStartPost` calls
`deploy/scripts/job-heartbeat.sh` to push
`shopass_job_last_success_unixtime{job_name=...}` to Pushgateway (observability
profile). Set `PUSHGATEWAY_URL` in `/etc/shopass/runtime.env` if not using
`http://127.0.0.1:9091`.

The supplied units assume the repository is checked out at `/srv/shopass` and
the root-owned runtime configuration is `/etc/shopass/runtime.env`. Adapt both
paths before installation.

```bash
sudo install -m 0644 deploy/systemd/shopass-scrape.service /etc/systemd/system/
sudo install -m 0644 deploy/systemd/shopass-scrape.timer /etc/systemd/system/
sudo install -m 0644 deploy/systemd/shopass-forecast.service /etc/systemd/system/
sudo install -m 0644 deploy/systemd/shopass-forecast.timer /etc/systemd/system/
sudo install -m 0644 deploy/systemd/shopass-backup.service /etc/systemd/system/
sudo install -m 0644 deploy/systemd/shopass-backup.timer /etc/systemd/system/
sudo install -m 0644 deploy/systemd/shopass-edge-attach.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now shopass-scrape.timer shopass-forecast.timer shopass-backup.timer
sudo systemctl enable --now shopass-edge-attach.service
systemctl list-timers 'shopass-*'
```

`shopass-backup.timer` runs `deploy/scripts/backup-pg.sh` at 03:15 (see
[`../RESTORE-RUNBOOK.md`](../RESTORE-RUNBOOK.md)). Before expecting off-host
DR, run [`../scripts/backup-upload-dry-run.sh`](../scripts/backup-upload-dry-run.sh)
and complete the VPS steps in [`../VPS-CHECKLIST.md`](../VPS-CHECKLIST.md).
`shopass-edge-attach.service` reconnects CyberOS Caddy to `shopass-edge` after
reboot if the CyberOS compose overlay is not yet installed.

`shopass-scrape.timer` runs every five minutes. `shopass-forecast.timer` runs
at 01:30 Asia/Ho_Chi_Minh, leaving a buffer before the 02:00 nightly scoring
cron that is already inside `dealsvc`. Do not add a second systemd timer for
`dealsvc` or it will duplicate scoring attempts.

Both services use `flock` and `--no-deps`: a missed/slow run does not overlap
with another run, and a timer never silently starts a partial application
stack. `Persistent=true` asks systemd to make up a missed calendar run after a
host outage; inspect the logs before treating that catch-up run as successful.

```bash
journalctl -u shopass-scrape.service -u shopass-forecast.service -f
sudo systemctl start shopass-scrape.service
```

The host runtime file currently contains transitional secret environment values
because the service binaries lack `_FILE`/Vault startup support. Follow
[`../secrets/README.md`](../secrets/README.md) and keep it mode `0600`.
