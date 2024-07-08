package models

type User struct {
	ID       string
	Login    string
	Email    string
	Password []byte
}
