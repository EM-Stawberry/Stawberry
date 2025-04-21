package product

type PostProductReq struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CategoryID  uint    `json:"category_id"`
	ShopID      uint    `json:"shop_id"`
	Price       float64 `json:"price"`
	Quantity    uint    `json:"quantity"`
}

type PostProductResp struct {
	ID uint `json:"id"`
}

type PatchProductReq struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	CategoryID  *uint    `json:"category_id,omitempty"`
	Quantity    *uint    `json:"quantity,omitempty"`
	ShopID      *uint    `json:"shop_id,omitempty"`
	ShopPointID *uint    `json:"shop_point_id,omitempty"`
}

type GetProductResp struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  uint   `json:"category_id"`
}
