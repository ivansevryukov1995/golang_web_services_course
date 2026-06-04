package dto

type User struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
}

type Article struct {
	Author         User     `json:"author"`
	Body           string   `json:"body"`
	Description    string   `json:"description"`
	Favorited      bool     `json:"favorited"`
	FavoritesCount int      `json:"favoritesCount"`
	Slug           string   `json:"slug" testdiff:"ignore"`
	TagList        []string `json:"tagList"`
	Title          string   `json:"title"`
	CreatedAt      string
	UpdatedAt      string
}

type ArticleRequest struct {
	Article Article `json:"article"`
}
