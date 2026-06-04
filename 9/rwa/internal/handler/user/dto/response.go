package dto

type User struct {
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Token     string `json:"token,omitempty"`
}

type UserResponse struct {
	User User `json:"user"`
}
