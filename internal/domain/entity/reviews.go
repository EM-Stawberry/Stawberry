package entity

import "time"

type ProductReview struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	UserID    int       `json:"user_id"`
	Rating    int       `json:"rating"`
	Review    string    `json:"review"`
	CreatedAt time.Time `json:"created_at"`
}

type SellerReview struct {
	ID        int       `json:"id"`
	SellerID  int       `json:"seller_id"`
	UserID    int       `json:"user_id"`
	Rating    int       `json:"rating"`
	Review    string    `json:"review"`
	CreatedAd time.Time `json:"created_at"`
}
