# Website

The MyGo Cloud Music homepage, built with TanStack Start and Solid. It's prerendered to static files at build time and deployed to Cloudflare (Workers static assets — no Worker script runs).

```bash
bun install          # at the repo root
bun run build        # outputs to website/dist/client
bun run deploy       # builds and deploys with wrangler
```
