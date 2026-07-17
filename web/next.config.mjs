/** @type {import('next').NextConfig} */
// API traffic stays same-origin and is routed by Caddy to the gateway. Do not
// add service-to-service rewrites here: doing so bypasses JWT verification and
// exposes private APIs through the public web process.
const nextConfig = {};

export default nextConfig;
