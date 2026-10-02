# ShopScout

ShopScout is a Go-powered product discovery dashboard. It scrapes product listings from the [WebScraper.io demo shop](https://webscraper.io/test-sites/e-commerce/static), then lets you search, filter, sort, add, edit, and remove products in a locally persisted catalog.

## Features

- Scrape laptops, tablets, and touchscreen phones, with up to 10 pages per run.
- Concurrent page fetching with upstream status checks, request timeouts, and bounded response sizes.
- Product cards with prices, descriptions, ratings, stock status, and links to the source listing.
- Search, category filters, price/rating sorting, catalog summary stats, and manual product CRUD.
- JSON-backed catalog storage and a responsive dashboard served directly by the Go application.
- Vercel Go runtime configuration and a health-check endpoint.

## Run locally

Requirements: Go 1.25 or newer.

```sh
go run ./cmd/server
```

Open `http://localhost:8080`. The catalog is saved to `data/products.json` by default. Set `PORT` to use another port, or `DATA_FILE` to select a different catalog file.

Run the test suite with:

```sh
go test ./...
```

## Deploy on Vercel

Import this repository into Vercel with the project root set to the repository root. [`vercel.json`](./vercel.json) selects Vercel's Go framework preset and builds the server from `cmd/server`; the Go server listens on Vercel's `PORT` environment variable. The embedded dashboard and API are served by the same Go process, so no separate frontend build is needed.

Vercel's filesystem is not durable between deployments or runtime instances. On Vercel, the app therefore keeps catalog changes in process memory; data can reset when an instance restarts, and separate instances may have different catalogs. For durable, shared product data, connect a persistent external database and configure an appropriate store before relying on the catalog.

See Vercel's [Go runtime documentation](https://vercel.com/docs/functions/runtimes/go) for current runtime requirements and deployment behavior.

This is a single-user demo and does not include authentication: anyone with the deployed URL can view and modify its catalog. Add access control before using it for a public or production-facing catalog.

## API

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/api/health` | Health check |
| `GET` | `/api/categories` | Scrapable shop categories |
| `GET` | `/api/products` | List products; supports `search`, `category`, and `sort` query parameters |
| `POST` | `/api/scrape` | Scrape and upsert a category (`{"category":"laptops","pages":1}`) |
| `POST` | `/api/products` | Add a product |
| `PUT` | `/api/products/{id}` | Update a product |
| `DELETE` | `/api/products/{id}` | Remove a product |

The scraper targets the public demo shop for development and demonstration. Check a site's terms and obtain permission before scraping another shop.
