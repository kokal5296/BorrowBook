package book_borrow

import "time"

// BookBorrow represents the borrowing record of a book by a user.
type BookBorrow struct {
	ID          int        `json:"id"`
	BookID      int        `json:"book_id" validate:"required"`
	UserID      int        `json:"user_id" validate:"required"`
	Borrow_date time.Time  `json:"borrow_date, omitempty"`
	Return_date *time.Time `json:"return_date, omitempty"`
}
