package entity

type Product struct {
	ID          uint   `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CategoryID  uint   `db:"category_id"`
}
