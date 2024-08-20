package models

type User struct {
	ID        string
	Login     string
	Email     string
	Password  []byte
	CreatedAt string
	UpdatedAt string
	Balance   float32
	Discount  int32
}
