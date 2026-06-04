package dto

type ArticleResponse struct {
	Article Article `json:"article"`
}

type ArticlesResponse struct {
	Articles      []Article `json:"articles"`
	ArticlesCount int       `json:"articlesCount"`
}
