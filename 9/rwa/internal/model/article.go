package model

type Article struct {
	Author         User     `json:"author"`
	Body           string   `json:"body"`
	Description    string   `json:"description"`
	Favorited      bool     `json:"favorited"`
	FavoritesCount int      `json:"favoritesCount"`
	Slug           string   `json:"slug" testdiff:"ignore"`
	TagList        []string `json:"tagList"`
	Title          string   `json:"title"`
}
