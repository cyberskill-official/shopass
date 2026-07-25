# Security Policy

## Supported versions

Security fixes are applied on the default branch of this repository. Pre-release
and experimental packages in this monorepo are unsupported unless explicitly
noted.

## Reporting a vulnerability

Please report security issues privately. Do not open a public GitHub issue for
vulnerabilities that could expose user data, payment flows, or authentication
bypass.

Preferred contact: maintainers listed in [CODEOWNERS](.github/CODEOWNERS)
(Stephen Cheng / CyberSkill maintainers), or email `zintaen@gmail.com`.

Include:

- Affected component (`web/`, `extension/`, `services/*`, etc.)
- Steps to reproduce
- Impact assessment (data exposure, privilege escalation, DoS)
- Any proof-of-concept that stays within responsible disclosure

We aim to acknowledge reports within 5 business days and to provide a status
update within 14 days. Please allow a reasonable window for a fix before any
public disclosure.

## Scope notes

- Payment IPN and billing paths (`services/bill/`) are especially sensitive.
- Extension host permissions and sync payloads must never carry marketplace
  session cookies or platform auth tokens.
- Fabricated or mock security scan output must never be treated as an audit.
