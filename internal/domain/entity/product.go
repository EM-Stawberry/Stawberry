package entity

type Product struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	CategoryID  uint                   `json:"category_id"`
	Attributes  map[string]interface{} `json:"product_attributes"`
}
