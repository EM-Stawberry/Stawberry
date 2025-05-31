package model

type ShopPoint struct {
	ID          uint   `db:"id"`
	ShopID      uint   `db:"shop_id"`
	Address     string `db:"Address"`
	PhoneNumber string `db:"phone_number"`
}
