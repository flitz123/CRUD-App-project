# ShopScout

ShopScout is a Go-powered product discovery dashboard. It scrapes product listings from the [WebScraper.io demo shop](https://webscraper.io/test-sites/e-commerce/static), then lets you search, filter, sort, add, edit, and remove products in a locally persisted catalog.

## Features

- Scrape laptops, tablets, and touchscreen phones, with up to 10 pages per run.
- Concurrent page fetching with upstream status checks, request timeouts, and bounded response sizes.
- Product cards with prices, descriptions, ratings, stock status, and links to the source listing.
- Search, category filters, price/rating sorting, catalog summary stats, and manual product CRUD.
- JSON-backed catalog storage and a responsive dashboard served directly by the Go application.
- Render Blueprint configuration and a health-check endpoint.

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

## Deploy on Render

Connect this repository to Render and select **Blueprint** deployment. Render reads [`render.yaml`](./render.yaml), builds the Go server, and uses `/api/health` for its health check. The service listens on Render's `PORT` environment variable.

Render's default filesystem is ephemeral, so the catalog resets when an instance is replaced. To retain catalog changes across deploys, attach a persistent disk mounted at `/var/data` and add the service environment variable `DATA_FILE=/var/data/products.json` in the Render dashboard.

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
