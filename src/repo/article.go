package repo

import "example.com/grpc/blog/src/models"

/*
ArticleRepo provides basic Repository Interface for dealing with Articles
*/
type ArticleRepo interface {

	// GetArticles returns a slice of Articles and error
	GetArticles() ([]models.Article, error)

	// AddArticle attempts to add an article
	// returns ID and error
	AddArticle(*models.Article) (id string, err error)

	// UpdateArticle attempts to update an article and returns an error
	UpdateArticle(*models.Article) error

	// DeleteArticle attempts to delete an Article by ID and returns an error
	DeleteArticle(id string) error

	// GetArticle attempts to get an Article and returns a ref to an Article and an error
	GetArticle(id string) (*models.Article, error)
}
