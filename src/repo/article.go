package repo

import (
	"context"

	"example.com/grpc/blog/src/models"
)

/*
ArticleRepo provides basic Repository Interface for dealing with Articles
*/
type ArticleRepo interface {

	// FillArticles populates the channel with Articles
	// It also accepts an "interrupt" channel which can interrupt read at any point
	FillArticles(context.Context, chan<- models.Article, <-chan struct{}) error

	// AddArticle attempts to add an article
	// returns ID and error
	AddArticle(context.Context, *models.Article) (string, error)

	// UpdateArticle attempts to update an article and returns an error
	UpdateArticle(context.Context, *models.Article) error

	// DeleteArticle attempts to delete an Article by ID and returns an error
	DeleteArticle(context.Context, string) error

	// GetArticle attempts to get an Article and returns a ref to an Article and an error
	GetArticle(context.Context, string) (*models.Article, error)
}
