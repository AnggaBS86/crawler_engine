# crawler_engine

Echo-based API + CLI for crawling a website and saving a rendered HTML snapshot (SPA/SSR/PWA).

Based on tutorial : https://medium.com/juliansplayground/crawling-dynamic-websites-using-chromedp-in-go-660feee9126a 

## Requirements

- Go 1.21+
- A Chromium/Chrome binary available on the machine (for SPA/PWA rendering via `chromedp`)

## Run API

```bash
go run ./cmd/server
```

Environment variables:

- `PORT` (default: `8080`)
- `OUTPUT_DIR` (default: `/media/user/New Volume/go/crawler_engine/output`)
- `CRAWL_TIMEOUT_SECONDS` (default: `45`)
- `CRAWL_NETWORK_IDLE_MS` (default: `500`) consider page "loaded" when network is idle
- `CRAWL_NETWORK_IDLE_MAX_MS` (default: `6000`) max time to wait for network idle (some sites never go idle)
- `CRAWL_WAIT_MS` (default: `2000`) extra wait after DOM ready (helps SPAs)
- `CHROME_BIN` (optional) override Chrome/Chromium binary (example: `google-chrome`)

Example request:

```bash
curl -sS -X POST localhost:8080/crawl \
  -H 'content-type: application/json' \
  -d '{"url":"https://cmlabs.co","filename":"cmlabs.co.html"}'
```

## Generate required snapshots (CLI)

```bash
go run ./cmd/crawl --url https://cmlabs.co --out snapshots/cmlabs.co.html
go run ./cmd/crawl --url https://sequence.day --out snapshots/sequence.day.html
go run ./cmd/crawl --url https://example.com --out snapshots/example.com.html
```

## Notes

- The crawler captures the rendered DOM (`document.documentElement.outerHTML`).
- This is intended to work for SPA/SSR/PWA, as long as the site can be rendered by headless Chromium.
- The API reuses a single headless Chrome instance to keep subsequent requests fast.

## Troubleshooting

If you see `chrome failed to start`, first verify Chrome can run headless on your machine:

```bash
google-chrome --headless --no-sandbox --disable-gpu --disable-dev-shm-usage --disable-features=Crashpad https://example.com
```
