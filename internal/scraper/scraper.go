package scraper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"shop-scraper/internal/product"
)

const sourceName = "WebScraper.io demo shop"

type Category struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

var categories = []Category{
	{ID: "laptops", Label: "Laptops", Path: "computers/laptops"},
	{ID: "tablets", Label: "Tablets", Path: "computers/tablets"},
	{ID: "phones", Label: "Touchscreen phones", Path: "phones/touch"},
}

type Scraper struct {
	client  *http.Client
	baseURL string
}

func New(client *http.Client, baseURL string) *Scraper {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	if baseURL == "" {
		baseURL = "https://webscraper.io/test-sites/e-commerce/static"
	}
	return &Scraper{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *Scraper) Categories() []Category {
	return append([]Category(nil), categories...)
}

func (s *Scraper) Scrape(ctx context.Context, categoryID string, pages int) ([]product.Product, error) {
	category, ok := categoryByID(categoryID)
	if !ok {
		return nil, fmt.Errorf("unknown category %q", categoryID)
	}
	if pages < 1 || pages > 10 {
		return nil, fmt.Errorf("pages must be between 1 and 10")
	}

	results := make([][]product.Product, pages)
	errs := make([]error, pages)
	var wg sync.WaitGroup
	for page := 1; page <= pages; page++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index], errs[index] = s.scrapePage(ctx, category, index+1)
		}(page - 1)
	}
	wg.Wait()

	var products []product.Product
	for i, pageProducts := range results {
		if errs[i] != nil {
			return nil, fmt.Errorf("scrape %s page %d: %w", category.Label, i+1, errs[i])
		}
		products = append(products, pageProducts...)
	}
	return products, nil
}

func (s *Scraper) scrapePage(ctx context.Context, category Category, page int) ([]product.Product, error) {
	target, err := url.Parse(s.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse scraper base URL: %w", err)
	}
	target = target.JoinPath(category.Path)
	if page > 1 {
		query := target.Query()
		query.Set("page", strconv.Itoa(page))
		target.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "ShopScout/1.0 (+https://webscraper.io/test-sites/e-commerce/static)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("parse page HTML: %w", err)
	}
	base, err := url.Parse(target.String())
	if err != nil {
		return nil, fmt.Errorf("parse page URL: %w", err)
	}

	products := make([]product.Product, 0)
	var parseErr error
	doc.Find(".thumbnail").Each(func(_ int, card *goquery.Selection) {
		link := card.Find("a.title").First()
		title := strings.TrimSpace(link.Text())
		if title == "" {
			title, _ = link.Attr("title")
			title = strings.TrimSpace(title)
		}
		if title == "" {
			return
		}
		price, err := parsePrice(card.Find(".price").Text())
		if err != nil {
			parseErr = fmt.Errorf("%s: %w", title, err)
			return
		}

		productURL := resolveURL(base, attr(link, "href"))
		imageURL := resolveURL(base, attr(card.Find("img").First(), "src"))
		rating, reviews := parseRating(card)
		products = append(products, product.Product{
			ID:          stableID(productURL, category.ID, title),
			Title:       title,
			Description: strings.TrimSpace(card.Find(".description").Text()),
			Price:       price,
			Category:    category.ID,
			Rating:      rating,
			Reviews:     reviews,
			Available:   !strings.Contains(strings.ToLower(card.Text()), "out of stock"),
			ImageURL:    imageURL,
			ProductURL:  productURL,
			Source:      sourceName,
			UpdatedAt:   time.Now().UTC(),
		})
	})
	if parseErr != nil {
		return nil, fmt.Errorf("parse product price: %w", parseErr)
	}
	if len(products) == 0 {
		return nil, fmt.Errorf("no products found; the shop page may have changed")
	}
	return products, nil
}

func categoryByID(id string) (Category, bool) {
	for _, category := range categories {
		if category.ID == id {
			return category, true
		}
	}
	return Category{}, false
}

func parsePrice(text string) (float64, error) {
	text = strings.NewReplacer("$", "", ",", "", " ", "").Replace(strings.TrimSpace(text))
	price, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(price) || math.IsInf(price, 0) || price < 0 {
		return 0, fmt.Errorf("invalid price %q", text)
	}
	return price, nil
}

func parseRating(card *goquery.Selection) (float64, int) {
	stars := card.Find(".ratings .glyphicon-star, .ratings .glyphicon-star-empty").Length()
	reviewsText := strings.TrimSpace(card.Find(".ratings .pull-right").Text())
	fields := strings.Fields(reviewsText)
	reviews := 0
	if len(fields) > 0 {
		reviews, _ = strconv.Atoi(fields[0])
	}
	return float64(stars), reviews
}

func attr(selection *goquery.Selection, key string) string {
	value, _ := selection.Attr(key)
	return strings.TrimSpace(value)
}

func resolveURL(base *url.URL, value string) string {
	if value == "" {
		return ""
	}
	ref, err := url.Parse(value)
	if err != nil {
		return ""
	}
	return base.ResolveReference(ref).String()
}

func stableID(productURL, category, title string) string {
	sum := sha256.Sum256([]byte(productURL + "|" + category + "|" + title))
	return hex.EncodeToString(sum[:12])
}
