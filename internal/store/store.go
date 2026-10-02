package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"shop-scraper/internal/product"
)

var ErrNotFound = errors.New("product not found")

type Store struct {
	mu       sync.RWMutex
	filePath string
	products map[string]product.Product
}

func New(filePath string) (*Store, error) {
	s := &Store{filePath: filePath, products: make(map[string]product.Product)}
	if filePath == "" {
		return s, nil
	}
	data, err := os.ReadFile(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read product catalog: %w", err)
	}
	var products []product.Product
	if err := json.Unmarshal(data, &products); err != nil {
		return nil, fmt.Errorf("decode product catalog: %w", err)
	}
	for _, item := range products {
		if item.ID == "" {
			return nil, fmt.Errorf("decode product catalog: product is missing an ID")
		}
		s.products[item.ID] = item
	}
	return s, nil
}

func (s *Store) List() []product.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	products := make([]product.Product, 0, len(s.products))
	for _, item := range s.products {
		products = append(products, item)
	}
	sort.Slice(products, func(i, j int) bool {
		return strings.ToLower(products[i].Title) < strings.ToLower(products[j].Title)
	})
	return products
}

func (s *Store) UpsertMany(items []product.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := clone(s.products)
	now := time.Now().UTC()
	for _, item := range items {
		if item.ID == "" {
			return fmt.Errorf("cannot save product without an ID")
		}
		item.UpdatedAt = now
		next[item.ID] = item
	}
	if err := s.persist(next); err != nil {
		return err
	}
	s.products = next
	return nil
}

func (s *Store) Create(item product.Product) (product.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := clone(s.products)
	if item.ID == "" {
		item.ID = nextLocalID(next)
	} else if _, exists := next[item.ID]; exists {
		return product.Product{}, fmt.Errorf("product ID already exists")
	}
	item.UpdatedAt = time.Now().UTC()
	next[item.ID] = item
	if err := s.persist(next); err != nil {
		return product.Product{}, err
	}
	s.products = next
	return item, nil
}

func (s *Store) Update(id string, item product.Product) (product.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := clone(s.products)
	existing, exists := next[id]
	if !exists {
		return product.Product{}, ErrNotFound
	}
	item.ID = id
	if item.Rating == 0 {
		item.Rating = existing.Rating
	}
	if item.Reviews == 0 {
		item.Reviews = existing.Reviews
	}
	if item.ProductURL == "" {
		item.ProductURL = existing.ProductURL
	}
	if item.Source == "" {
		item.Source = existing.Source
	}
	item.UpdatedAt = time.Now().UTC()
	next[id] = item
	if err := s.persist(next); err != nil {
		return product.Product{}, err
	}
	s.products = next
	return item, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := clone(s.products)
	if _, exists := next[id]; !exists {
		return ErrNotFound
	}
	delete(next, id)
	if err := s.persist(next); err != nil {
		return err
	}
	s.products = next
	return nil
}

func (s *Store) persist(products map[string]product.Product) error {
	if s.filePath == "" {
		return nil
	}
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create catalog directory: %w", err)
	}
	items := make([]product.Product, 0, len(products))
	for _, item := range products {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("encode product catalog: %w", err)
	}
	temp, err := os.CreateTemp(dir, ".products-*.json")
	if err != nil {
		return fmt.Errorf("create catalog temp file: %w", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("write product catalog: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("sync product catalog: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close product catalog: %w", err)
	}
	if err := os.Rename(tempName, s.filePath); err != nil {
		return fmt.Errorf("replace product catalog: %w", err)
	}
	return nil
}

func clone(items map[string]product.Product) map[string]product.Product {
	next := make(map[string]product.Product, len(items))
	for id, item := range items {
		next[id] = item
	}
	return next
}

func nextLocalID(items map[string]product.Product) string {
	for id := 1; ; id++ {
		candidate := fmt.Sprintf("local-%d", id)
		if _, exists := items[candidate]; !exists {
			return candidate
		}
	}
}
