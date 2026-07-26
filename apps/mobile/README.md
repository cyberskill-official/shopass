# Shopass mobile (scaffold)

TypeScript modules for React Native (MOBILE-001/002/003):

- Secure refresh storage contract + in-memory access token
- Gateway HTTP client with single refresh retry
- FCM device register/unregister → `/v1/devices`
- Track/chart + explicit-tap cart optimize clients
- Share/deeplink helpers (`product_id` + `ref` only) with pending referral + anti-self-referral

Full RN app shell (Expo/CLI), Keychain bindings, and AASA/assetlinks are follow-ups once the scaffold lands.

```bash
cd apps/mobile && npm ci && npm test
```
