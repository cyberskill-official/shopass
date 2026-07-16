# Path Layout & Least Privilege Policy (TASK-INFRA-003)

## Path Layout
Secrets are isolated into logical namespaces using path prefixes. This guarantees separation of concerns and limits blast radius in case of a service compromise.

- `db/main`
  - Purpose: Credentials for the primary PostgreSQL database.
  - Allowed Readers: Any backend service that needs direct database access.

- `auth/jwt-signing`
  - Purpose: The RS256 private key used to sign JWTs.
  - Allowed Readers: `auth` service only.

- `scrape/proxy/brightdata`
  - Purpose: Proxy credentials for Bright Data vendor.
  - Allowed Readers: `scrape` service only.

- `scrape/proxy/oxylabs`
  - Purpose: Proxy credentials for Oxylabs vendor.
  - Allowed Readers: `scrape` service only.

- `gateway/tokens`
  - Purpose: Internal tokens for edge-to-service communication if needed.
  - Allowed Readers: `gateway` service only.

- `bill/momo` and `bill/zalopay`
  - Purpose: Payment gateway keys.
  - Allowed Readers: `bill` service only.

## Least Privilege Policy
- Services must be explicitly granted `read` permission only to the exact paths they require.
- No service is permitted to perform `list` or `read` on the root secret path `*`.
- When using Vault, this translates to specific HCL policies tied to the service's AppRole.
- When using AWS Secrets Manager, this translates to IAM Roles for Service Accounts (IRSA) with strictly scoped `secretsmanager:GetSecretValue` permissions specifying resource ARNs.
