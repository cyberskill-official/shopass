# Production secret contract

`docker-compose.production.yml` accepts a Docker Compose secret file for the
initial Postgres password:

```
sudo install -d -m 0700 /etc/shopass/secrets
sudo sh -c 'umask 077; openssl rand -base64 36 > /etc/shopass/secrets/db_password'
```

Set `DB_PASSWORD_FILE=/etc/shopass/secrets/db_password` in the root-owned
runtime environment file. The database receives this as
`/run/secrets/db_password` through `POSTGRES_PASSWORD_FILE`; the password is
not committed and is not exposed as a Compose variable.

Generate an auth-only RS256 signing key and mount it only into `authsvc`:

```
sudo sh -c 'umask 077; openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:3072 > /etc/shopass/secrets/auth_signing_key.pem'
```

Set `AUTH_SIGNING_KEY_SECRET_FILE` to that path and set a stable, non-secret
`AUTH_KEY_ID` in the runtime environment. `authsvc` receives the PEM at
`/run/secrets/auth_signing_key` and fails to start when it is absent in
`APP_ENV=production`; it does not generate a new production key at boot.

## Current application limitation

The Go, Node, and Python services currently only read `DATABASE_URL` (and some
third-party credentials) from environment variables. They do **not** implement
`DATABASE_URL_FILE` or use the repository's Vault/AWS secret provider at
startup. Therefore an operator must temporarily inject the following values
into `/etc/shopass/runtime.env` (mode `0600`, root-owned) or an equivalent
platform secret injector:

- `DATABASE_URL`
- `GOOGLE_CLIENT_SECRET`, only if Google OAuth is enabled
- `HTTPS_PROXY`, when it contains proxy credentials
- `PRICE_INTERNAL_SERVICE_TOKEN`, a randomly generated shared value supplied
  identically to `tracksvc` and `pricesvc`. It authenticates private
  track-to-price calls and pricesvc→tracksvc price-changed notifications
  through `X-Service-Token`; do not put it in Caddy, browser configuration,
  logs, or a source-controlled env file.
- `BILL_INTERNAL_SERVICE_TOKEN`, shared by tracksvc and billsvc for gating.
- `COMPLY_OPERATOR_TOKEN`, required for breach open/advance/close on the
  private complysvc network. The public gateway does not allowlist those
  routes. Caddy and the gateway strip `X-Operator-Token` from browsers.

This is a transitional compatibility requirement, not a claim that the
application meets FR-INFRA-003's no-secret-in-environment invariant. Before a
public production launch, add source-level file/Vault loading and remove those
environment values.

## Required additions before declaring secrets complete

1. Make each service load `DATABASE_URL_FILE` and scoped private-service
   credentials (or use Vault/AWS secrets) before opening its database pool or
   making authenticated service-to-service calls.
2. Add controlled key-rotation overlap for `authsvc`: publish the previous
   public JWKS key during the access-token grace period, then retire it by an
   audited operator procedure.
3. Give the gateway only JWKS/public-key access and its own Redis credentials;
   it must never receive the auth signing private key.
4. Rotate and test restore of database, OAuth, proxy, notification, billing,
   and signing secrets. Do not log secret values.
