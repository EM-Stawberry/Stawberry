package entity

type Product struct {
	ID            int                    `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	CategoryID    uint                   `json:"category_id"`
	MinimalPrice  float64                `json:"minimal_price"`
	MaximalPrice  float64                `json:"maximal_price"`
	AverageRating float64                `json:"average_rating"`
	CountReviews  int                    `json:"count_reviews"`
	Attributes    map[string]interface{} `json:"product_attributes"`
}

type NewProduct struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  int    `json:"category_id"`
}
