package repo

import (
	"context"
	"fmt"
	"log"
	"sync"

	"example.com/grpc/blog/src/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
ArticleRepo provides basic Repository Interface for dealing with Articles
*/
type ArticleRepo interface {

	// FillArticles populates the channel with Articles
	// It also accepts "stop" channel which should be called explicitly
	FillArticles(context.Context, chan<- models.Article, <-chan struct{}) error

	// AddArticle attempts to add an article
	// returns ID and error
	AddArticle(context.Context, *models.Article) (string, error)

	// UpdateArticle attempts to update an article and returns an error
	UpdateArticle(context.Context, *models.Article) (*models.Article, error)

	// DeleteArticle attempts to delete an Article by ID and returns deleted Article and an error
	DeleteArticle(context.Context, string) (*models.Article, error)

	// GetArticle attempts to get an Article and returns a ref to an Article and an error
	GetArticle(context.Context, string) (*models.Article, error)
}

// MapArticleRepo is used for testing (or in-memory storage for Articles)
type MapArticleRepo struct {
	articles map[primitive.ObjectID]models.Article
}

// NewMapRepo creates a struct literal of Map Repo and returns a pointer to it
func NewMapRepo(m map[primitive.ObjectID]models.Article) *MapArticleRepo {
	return &MapArticleRepo{
		articles: m,
	}
}

// AddArticle to the map
func (m *MapArticleRepo) AddArticle(ctx context.Context, a *models.Article) (string, error) {
	id := primitive.NewObjectID()
	a.ID = id
	m.articles[id] = *a
	return id.String(), nil
}

// GetArticle from the map
func (m *MapArticleRepo) GetArticle(ctx context.Context, id string) (*models.Article, error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	a, ok := m.articles[oid]
	if !ok {
		return nil, fmt.Errorf("Missing value for %v", id)
	}
	return &a, nil
}

// DeleteArticle from the map
func (m *MapArticleRepo) DeleteArticle(ctx context.Context, id string) (*models.Article, error) {
	oid, _ := primitive.ObjectIDFromHex(id)
	a, ok := m.articles[oid]
	if !ok {
		return nil, fmt.Errorf("Missing value for %v", id)
	}
	delete(m.articles, oid)
	return &a, nil
}

// UpdateArticle inside the map
func (m *MapArticleRepo) UpdateArticle(ctx context.Context, a *models.Article) (*models.Article, error) {
	ua, ok := m.articles[a.ID]
	if !ok {
		return nil, fmt.Errorf("Missing old model %v", a)
	}
	m.articles[a.ID] = ua
	return a, nil
}

// FillArticles from the map
func (m *MapArticleRepo) FillArticles(ctx context.Context, out chan<- models.Article, stop <-chan struct{}) error {
	defer close(out)
	// get map keys to traverse 1 by 1 later
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for _, v := range m.articles {
			select {
			case <-stop:
				// stop signal
				wg.Done()
				return
			default:
				out <- v
			}
		}
		wg.Done()
	}()

	go func() {
		for {
			select {
			case <-stop:
				log.Print("got stop signal after loop")
				return
			}
		}
	}()

	wg.Wait()
	return nil
}
