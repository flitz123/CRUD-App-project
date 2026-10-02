package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScrapeParsesProductsAndCategoryPages(t *testing.T) {
	requests := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.URL.Path + "?" + r.URL.RawQuery
		fmt.Fprint(w, `<div class="thumbnail">
			<a class="title" href="/product/1" title="Everyday laptop">Laptop</a>
			<img src="/images/laptop.jpg">
			<h4 class="price">$1,249.50</h4>
			<p class="description">A useful laptop</p>
			<div class="ratings"><span class="glyphicon-star"></span><span class="glyphicon-star"></span><span class="pull-right">18 reviews</span></div>
		</div>`)
	}))
	defer server.Close()

	items, err := New(server.Client(), server.URL).Scrape(context.Background(), "laptops", 2)
	if err != nil {
		t.Fatalf("Scrape returned an error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("Scrape returned %d products, want 2", len(items))
	}
	item := items[0]
	if item.Title != "Laptop" || item.Price != 1249.5 || item.Category != "laptops" {
		t.Fatalf("unexpected product fields: %+v", item)
	}
	if item.Rating != 2 || item.Reviews != 18 {
		t.Fatalf("unexpected rating fields: %+v", item)
	}
	if item.ProductURL != server.URL+"/product/1" || item.ImageURL != server.URL+"/images/laptop.jpg" {
		t.Fatalf("product links were not resolved: %+v", item)
	}
	if len(requests) != 2 {
		t.Fatalf("received %d page requests, want 2", len(requests))
	}
}

func TestScrapeRejectsInvalidCategoryAndPageCount(t *testing.T) {
	s := New(nil, "http://example.test")
	if _, err := s.Scrape(context.Background(), "books", 1); err == nil {
		t.Fatal("Scrape accepted an unknown category")
	}
	if _, err := s.Scrape(context.Background(), "laptops", 0); err == nil {
		t.Fatal("Scrape accepted a zero page count")
	}
}

func TestScrapeReportsUpstreamStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	if _, err := New(server.Client(), server.URL).Scrape(context.Background(), "tablets", 1); err == nil {
		t.Fatal("Scrape ignored an upstream error")
	}
}
