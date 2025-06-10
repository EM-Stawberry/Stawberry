package model

type ShopPointInventory struct {
	ShopPointID uint    `db:"shop_point_id"`
	ProductID   uint    `db:"product_id"`
	Price       float64 `db:"price"`
	Quantity    uint    `db:"quantity"`
}
