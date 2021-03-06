package repo

import (
	"context"
	"fmt"
	"log"
	"os"

	"example.com/grpc/blog/src/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func (r *MongoArticleRepo) AddArticle(ctx context.Context, a *models.Article) (id string, err error) {
	res, err := r.c.InsertOne(ctx, a)
	if err != nil {
		return "", err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}
	return "", fmt.Errorf("Got wrong type for Mongo Object ID")
}

// FillArticles graps documents from MongoDB and sends to "out" channel
func (r *MongoArticleRepo) FillArticles(ctx context.Context, out chan<- models.Article, stop <-chan struct{}) error {
	defer close(out)
	c, err := r.c.Find(ctx, bson.D{})
	if err != nil {
		return err
	}
	e := make(chan error)

	go func() {
		defer close(e)
		for c.Next(ctx) {
			select {
			case <-stop:
				fmt.Println("Interrupt signal received, cancelling read")
				e <- nil
				return
			default:
				m := models.Article{}
				err := c.Decode(&m)
				if err != nil {
					e <- err
					return
				}
				out <- m
			}
		}
		if err = c.Err(); err != nil {
			e <- err
		}
		e <- nil
	}()

	// We need to make sure stop can be cought after Loop is over
	go func() {
		for {
			select {
			case <-stop:
				log.Print("got stop signal after loop")
				return
			}
		}
	}()

	return <-e
}

// UpdateArticle attempts to update an article
func (r *MongoArticleRepo) UpdateArticle(ctx context.Context, a *models.Article) (*models.Article, error) {
	res := r.c.FindOneAndUpdate(ctx, bson.M{"_id": a.ID}, bson.M{"$set": a}, options.FindOneAndUpdate().SetReturnDocument(options.After))
	m := models.Article{}
	err := res.Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// DeleteArticle attempts to delete article by object id
func (r *MongoArticleRepo) DeleteArticle(ctx context.Context, id string) (*models.Article, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	res := r.c.FindOneAndDelete(ctx, bson.M{"_id": oid})
	m := models.Article{}
	err = res.Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetArticle gets an article by ID
func (r *MongoArticleRepo) GetArticle(ctx context.Context, id string) (*models.Article, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	res := r.c.FindOne(ctx, bson.M{"_id": oid})
	m := models.Article{}
	err = res.Decode(&m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
