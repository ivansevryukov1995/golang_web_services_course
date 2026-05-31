package model

type User struct {
	ID        string
	Email     string
	Password  string
	Username  string
	Bio       string
	Image     string
	Following bool
	CreatedAt string
	UpdatedAt string
}
