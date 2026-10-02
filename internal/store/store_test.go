package store

import (
	"errors"
	"testing"

	"shop-scraper/internal/product"
)

func TestProductChangesPersistAcrossStoreInstances(t *testing.T) {
	path := t.TempDir() + "/products.json"
	catalog, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	created, err := catalog.Create(product.Product{Title: "Laptop", Category: "laptops", Price: 799})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "local-1" {
		t.Fatalf("created ID = %q, want local-1", created.ID)
	}
	created.Price = 749
	if _, err := catalog.Update(created.ID, created); err != nil {
		t.Fatal(err)
	}
	reloaded, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	items := reloaded.List()
	if len(items) != 1 || items[0].Price != 749 {
		t.Fatalf("reloaded products = %+v, want updated laptop", items)
	}
	if err := reloaded.Delete(created.ID); err != nil {
		t.Fatal(err)
	}
	if err := reloaded.Delete(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleting a missing product returned %v, want ErrNotFound", err)
	}
}

func TestUpsertManyUpdatesExistingProduct(t *testing.T) {
	catalog, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	item := product.Product{ID: "remote-1", Title: "Tablet", Category: "tablets", Price: 499, Rating: 4, Reviews: 12, ProductURL: "https://example.test/tablet", Source: "Demo shop"}
	if err := catalog.UpsertMany([]product.Product{item}); err != nil {
		t.Fatal(err)
	}
	item.Price = 449
	if err := catalog.UpsertMany([]product.Product{item}); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Update(item.ID, product.Product{
		Title: "Tablet", Description: "Updated details", Category: "tablets", Price: 449, Available: true,
	}); err != nil {
		t.Fatal(err)
	}
	items := catalog.List()
	if len(items) != 1 || items[0].Price != 449 || items[0].Rating != 4 || items[0].Reviews != 12 || items[0].ProductURL == "" || items[0].Source == "" {
		t.Fatalf("upserted products = %+v, want one updated product", items)
	}
}
