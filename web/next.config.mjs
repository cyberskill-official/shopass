/** @type {import('next').NextConfig} */
// Browser /v1/* traffic is same-origin. Production Caddy intercepts /v1 before
// Next.js and sends it to the JWT gateway. Local Docker without Caddy uses the
// App Router catch-all at app/v1/[...path] which proxies only to
// GATEWAY_INTERNAL_BASE_URL (never to pricesvc/tracksvc). Do not add
// service-to-service rewrites here: that would bypass JWT verification.
const securityHeaders = [
  { key: "X-Content-Type-Options", value: "nosniff" },
  { key: "X-Frame-Options", value: "DENY" },
  { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
  {
    key: "Permissions-Policy",
    value: "camera=(), microphone=(), geolocation=(), payment=()",
  },
  {
    key: "Content-Security-Policy",
    value: [
      "default-src 'self'",
      "base-uri 'self'",
      "frame-ancestors 'none'",
      "form-action 'self'",
      "img-src 'self' data: https:",
      "font-src 'self' data:",
      // Next.js requires inline scripts/styles in the App Router; tighten later
      // if a nonce-based CSP pipeline is added.
      "script-src 'self' 'unsafe-inline' 'unsafe-eval'",
      "style-src 'self' 'unsafe-inline'",
      "connect-src 'self'",
      "object-src 'none'",
    ].join("; "),
  },
];

const nextConfig = {
  compress: true,
  images: {
    formats: ["image/avif", "image/webp"],
    remotePatterns: [
      { protocol: "https", hostname: "shopass.cyberskill.world" },
    ],
  },
  async redirects() {
    return [
      // Spec R44 used the old product name in the slug; keep SEO alias.
      {
        source: "/so-sanh/sandeal-vs-beecost",
        destination: "/so-sanh/shopass-vs-beecost",
        permanent: true,
      },
    ];
  },
  async headers() {
    return [
      {
        source: "/:path*",
        headers: securityHeaders,
      },
    ];
  },
};

export default nextConfig;
