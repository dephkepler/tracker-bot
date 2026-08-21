import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // Static export: every page is a client component talking to the Go API with
  // a per-launch credential the server cannot see, so there is nothing for a
  // Node runtime to render. Caddy serves the files; no second container, no
  // runtime to keep patched, and the whole bundle is immutable-cacheable —
  // which matters because a Mini App cold-starts in a WebView on mobile data
  // every time it opens.
  output: 'export',
  // Emits out/track/index.html, so Caddy resolves /track with no redirect hop.
  trailingSlash: true,
}

export default nextConfig
