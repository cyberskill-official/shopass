# Forecast UX — cold-start vs mature SKU

Dealsvc nightly scoring only fires `bottom_predicted` alerts for SKUs that pass
the maturity gate (`IsFeatureReady` / TASK-DEAL-002). Cold-start SKUs stay
silent until enough price history exists.

| State | Meaning | User-facing |
|-------|---------|-------------|
| Cold-start | Insufficient history / not mature | Chart may show sparse series; no `bottom_predicted` push |
| Mature | Gate passes; `price_forecast.p_bottom_14d` scored | When `p_bottom_14d > 0.7` and an active alert rule exists, dealsvc enqueues notif |

Smoke / staging injects a mature forecast row (`p_bottom_14d = 0.85`) so the
deal→notif path can be proven without waiting for Prophet/CmdStan. Production
mlforecast jobs write the same columns for live SKUs.

See also: [`docs/STAGING-MONETIZED-SLICE.md`](STAGING-MONETIZED-SLICE.md) Path 1.
