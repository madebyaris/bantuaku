# Google Trends Scraper

Headless browser scraper for Google Trends data (no official API).

## Overview

This tool scrapes Google Trends data including:
- Interest over time (time series)
- Related queries (rising, top, related)

## Usage

```bash
# Build
go build -o trends-scraper scraper.go

# Run
./trends-scraper --keyword "tokopedia" --geo "ID"
```

## Features

- Headless browser automation (chromedp)
- Retry with exponential backoff
- Rate limiting compliance
- Time series extraction
- Related queries capture

## Important Notes

⚠️ **Terms of Use Compliance**
- Respect Google's robots.txt
- Implement rate limiting (delay between requests)
- Use responsibly and ethically
- Consider caching aggressively
- Monitor for IP blocks

## Dependencies

- chromedp (headless Chrome automation)
- Go 1.22+

## Configuration

Set environment variables:
- `TRENDS_RATE_LIMIT_DELAY` - Milliseconds between requests (default: 2000)
- `TRENDS_MAX_RETRIES` - Maximum retry attempts (default: 3)

## Integration

This scraper is called by the backend ingestion service (`backend/services/trends/ingest.go`).

