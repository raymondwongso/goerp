package domain

// Example is a temporary struct used by example module. It doesn't serve any real purpose
type Example struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}
