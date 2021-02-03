package repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"example.com/grpc/blog/src/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoArticleRepo is the Article repository implementation in MongoDB
type MongoArticleRepo struct {
	c *mongo.Collection
}

// NewMongoArticleRepo returns initialized MongoDB article repo
func NewMongoArticleRepo(c *mongo.Client) *MongoArticleRepo {
	return &MongoArticleRepo{
		c: c.Database(os.Getenv("DB")).Collection("articles"),
	}
}

// AddArticle implements ArticleRepo.AddArticle by persisting articles in MongoDB
func (r *MongoArticleRepo) AddArticle(a *models.Article) (id string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2*time.Second))
	defer cancel()
	res, err := r.c.InsertOne(ctx, a)
	if err != nil {
		return "", err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "", fmt.Errorf("Got wrong type for Mongo Object ID")
}

// GetArticles returns a slice of Article models from articles collection
func (r *MongoArticleRepo) GetArticles() ([]models.Article, error) {
	return nil, errors.New("not implemented")
}

// UpdateArticle attempts to update an article
func (r *MongoArticleRepo) UpdateArticle(a *models.Article) error {
	return nil
}

// DeleteArticle attempts to delete article by object id
func (r *MongoArticleRepo) DeleteArticle(id string) error {
	return nil
}

// GetArticle gets an article by ID
func (r *MongoArticleRepo) GetArticle(id string) (*models.Article, error) {
	return nil, errors.New("not implemented")
}
