package dto

import (
	"rwa/internal/model"
)

func ToArticleResponse(u model.Article) ArticleResponse {
	return ArticleResponse{
		Article: Article{
			Author: User{
				Bio:      u.Author.Bio,
				Username: u.Author.Username,
			},
			Body:           u.Body,
			Description:    u.Description,
			Favorited:      u.Favorited,
			FavoritesCount: u.FavoritesCount,
			Slug:           u.Slug,
			TagList:        u.TagList,
			Title:          u.Title,
			CreatedAt:      u.CreatedAt,
			UpdatedAt:      u.UpdatedAt,
		},
	}
}

func ToArticle(u model.Article) Article {
	return Article{
		Author: User{
			Bio:      u.Author.Bio,
			Username: u.Author.Username,
		},
		Body:           u.Body,
		Description:    u.Description,
		Favorited:      u.Favorited,
		FavoritesCount: u.FavoritesCount,
		Slug:           u.Slug,
		TagList:        u.TagList,
		Title:          u.Title,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

func ToArticlesResponse(articles []model.Article) ArticlesResponse {
	var items []Article

	for _, article := range articles {
		items = append(items, ToArticle(article))
	}

	return ArticlesResponse{
		Articles:      items,
		ArticlesCount: len(items),
	}
}

func ToArticleInput(req ArticleRequest) model.Article {
	return model.Article{
		Title:       req.Article.Title,
		Body:        req.Article.Body,
		Description: req.Article.Description,
		TagList:     req.Article.TagList,
	}
}
