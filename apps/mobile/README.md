# Shopass mobile (React Native TypeScript)

Logic-first scaffold for TASK-MOBILE-001/002/003:

- Auth via shared JWT (AUTH-002) with refresh in Keychain/Keystore
- FCM device registration (`POST /v1/devices`) + unregister on logout
- Thin track/checkout clients (no client-side sale math; user-initiated optimize only)
- Universal/app link share + pending referral attribution

```bash
npm install
npm test
```

Native RN shell / Firebase wiring is intentionally deferred until store credentials land; unit tests cover contracts with mocks.
