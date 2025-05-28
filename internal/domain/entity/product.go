package entity

type Product struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	CategoryID  uint                   `json:"category_id"`
	Attributes  map[string]interface{} `json:"product_attributes"`
}

type NewProduct struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  int    `json:"category_id"`
}
