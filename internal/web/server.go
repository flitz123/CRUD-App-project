package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"sort"
	"strings"

	"shop-scraper/internal/product"
	"shop-scraper/internal/scraper"
	"shop-scraper/internal/store"
)

//go:embed static
var staticFiles embed.FS

type Server struct {
	catalog *store.Store
	scraper *scraper.Scraper
	mux     *http.ServeMux
}

func New(catalog *store.Store, scraperClient *scraper.Scraper) *Server {
	s := &Server{catalog: catalog, scraper: scraperClient, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/categories", s.categories)
	s.mux.HandleFunc("GET /api/products", s.products)
	s.mux.HandleFunc("POST /api/products", s.createProduct)
	s.mux.HandleFunc("PUT /api/products/{id}", s.updateProduct)
	s.mux.HandleFunc("DELETE /api/products/{id}", s.deleteProduct)
	s.mux.HandleFunc("POST /api/scrape", s.scrape)
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(fmt.Sprintf("load embedded web assets: %v", err))
	}
	s.mux.Handle("GET /", http.FileServer(http.FS(static)))
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	s.mux.ServeHTTP(w, r)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) categories(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.scraper.Categories())
}

func (s *Server) products(w http.ResponseWriter, r *http.Request) {
	search := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	sortBy := r.URL.Query().Get("sort")
	products := s.catalog.List()
	filtered := make([]product.Product, 0, len(products))
	for _, item := range products {
		if category != "" && category != "all" && item.Category != category {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(item.Title+" "+item.Description), search) {
			continue
		}
		filtered = append(filtered, item)
	}
	switch sortBy {
	case "price-asc":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Price < filtered[j].Price })
	case "price-desc":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Price > filtered[j].Price })
	case "rating":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Rating > filtered[j].Rating })
	}
	writeJSON(w, http.StatusOK, filtered)
}

func (s *Server) createProduct(w http.ResponseWriter, r *http.Request) {
	var item product.Product
	if err := decodeJSON(w, r, &item); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateProduct(item, s.scraper.Categories()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if item.Source == "" {
		item.Source = "Manual entry"
	}
	created, err := s.catalog.Create(item)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) updateProduct(w http.ResponseWriter, r *http.Request) {
	var item product.Product
	if err := decodeJSON(w, r, &item); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateProduct(item, s.scraper.Categories()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := s.catalog.Update(r.PathValue("id"), item)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Product not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteProduct(w http.ResponseWriter, r *http.Request) {
	if err := s.catalog.Delete(r.PathValue("id")); errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "Product not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) scrape(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Category string `json:"category"`
		Pages    int    `json:"pages"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if request.Pages < 1 || request.Pages > 10 {
		writeError(w, http.StatusBadRequest, "Pages must be between 1 and 10")
		return
	}
	validCategory := false
	for _, category := range s.scraper.Categories() {
		if category.ID == request.Category {
			validCategory = true
			break
		}
	}
	if !validCategory {
		writeError(w, http.StatusBadRequest, "Choose a valid product category")
		return
	}
	products, err := s.scraper.Scrape(r.Context(), request.Category, request.Pages)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if err := s.catalog.UpsertMany(products); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":    len(products),
		"category": request.Category,
		"products": products,
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("Invalid request body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("Request body must contain one JSON value")
	}
	return nil
}

func validateProduct(item product.Product, categories []scraper.Category) error {
	if strings.TrimSpace(item.Title) == "" {
		return fmt.Errorf("Product name is required")
	}
	if item.Price < 0 {
		return fmt.Errorf("Price must be zero or greater")
	}
	for _, category := range categories {
		if category.ID == item.Category {
			return nil
		}
	}
	return fmt.Errorf("Choose a valid product category")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
