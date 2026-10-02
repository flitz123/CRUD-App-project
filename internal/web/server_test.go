package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"shop-scraper/internal/product"
	"shop-scraper/internal/scraper"
	"shop-scraper/internal/store"
)

func TestProductCRUDAndCategoryFiltering(t *testing.T) {
	catalog, err := store.New("")
	if err != nil {
		t.Fatal(err)
	}
	handler := New(catalog, scraper.New(nil, ""))

	created := performJSON(t, handler, http.MethodPost, "/api/products", map[string]any{
		"title": "Compact tablet", "category": "tablets", "price": 189.5, "available": true,
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body: %s", created.Code, created.Body)
	}
	var item product.Product
	if err := json.Unmarshal(created.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	if item.ID == "" {
		t.Fatal("create response has no ID")
	}

	updated := performJSON(t, handler, http.MethodPut, "/api/products/"+item.ID, map[string]any{
		"title": "Compact tablet", "category": "tablets", "price": 169, "available": false,
	})
	if updated.Code != http.StatusOK {
		t.Fatalf("update status = %d, body: %s", updated.Code, updated.Body)
	}
	filtered := httptest.NewRecorder()
	handler.ServeHTTP(filtered, httptest.NewRequest(http.MethodGet, "/api/products?category=tablets&search=compact", nil))
	if filtered.Code != http.StatusOK {
		t.Fatalf("filter status = %d", filtered.Code)
	}
	var items []product.Product
	if err := json.Unmarshal(filtered.Body.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Price != 169 || items[0].Available {
		t.Fatalf("filtered products = %+v", items)
	}

	deleted := httptest.NewRecorder()
	handler.ServeHTTP(deleted, httptest.NewRequest(http.MethodDelete, "/api/products/"+item.ID, nil))
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, body: %s", deleted.Code, deleted.Body)
	}
}

func TestCreateRejectsInvalidProduct(t *testing.T) {
	catalog, err := store.New("")
	if err != nil {
		t.Fatal(err)
	}
	handler := New(catalog, scraper.New(nil, ""))
	response := performJSON(t, handler, http.MethodPost, "/api/products", map[string]any{
		"title": "", "category": "books", "price": -1,
	})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid product status = %d, want 400", response.Code)
	}
}

func TestScrapeRejectsInvalidInputBeforeFetching(t *testing.T) {
	catalog, err := store.New("")
	if err != nil {
		t.Fatal(err)
	}
	handler := New(catalog, scraper.New(nil, ""))
	response := performJSON(t, handler, http.MethodPost, "/api/scrape", map[string]any{
		"category": "books", "pages": 1,
	})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid scrape status = %d, want 400", response.Code)
	}
}

func performJSON(t *testing.T, handler http.Handler, method, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
