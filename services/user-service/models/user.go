package models

type User struct {
	UserID       int
	Username     string
	Email        string
	PasswordHash string
}
