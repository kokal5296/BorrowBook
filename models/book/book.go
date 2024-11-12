package book

// Book represents a book available in the library.
type Book struct {
	ID       int    `json:"id"`
	Title    string `json:"title" validate:"required"`
	Quantity int    `json:"quantity" validate:"required"`
}
