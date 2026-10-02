package product

import "time"

type Product struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	Rating      float64   `json:"rating"`
	Reviews     int       `json:"reviews"`
	Available   bool      `json:"available"`
	ImageURL    string    `json:"image_url"`
	ProductURL  string    `json:"product_url"`
	Source      string    `json:"source"`
	UpdatedAt   time.Time `json:"updated_at"`
}
