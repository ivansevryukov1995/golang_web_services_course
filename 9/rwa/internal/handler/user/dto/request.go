package dto

type UserRequest struct {
	User struct {
		Email    string `json:"email"`
		Password string `json:"password,omitempty"`
		Username string `json:"username,omitempty"`
		Bio      string `json:"bio,omitempty"`
	} `json:"user"`
}
