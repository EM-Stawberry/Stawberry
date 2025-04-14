package entity

import "time"

type ProductReview struct {
	ID         int
	product_id int
	user_id    int
	rating     int
	review     string
	created_at time.Time
}
