package product

type Product struct {
	ID          uint
	Name        string
	Description string
	CategoryID  uint
	ShopID      uint
	Price       float64
	Quantity    uint
	ShopPointID uint
	ShopName    string
}

type UpdateProduct struct {
	Name        *string
	Description *string
	CategoryID  *uint
	Price       *float64
	Quantity    *uint
	ShopPointID *uint
}
