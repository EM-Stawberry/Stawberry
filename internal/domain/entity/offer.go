package entity

import "time"

type Offer struct {
	ID        uint      `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	ProductID uint      `json:"product_id" db:"product_id"`
	ShopID    uint      `json:"store_id" db:"shop_id"`
	Price     float64   `json:"price" db:"offer_price"`
	Status    string    `json:"status" db:"status"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
