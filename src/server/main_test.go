package server

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	pb "example.com/grpc/blog/gen/src"
	"example.com/grpc/blog/src/models"
	"example.com/grpc/blog/src/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreate_success(t *testing.T) {
	r := &pb.CreateRequest{
		Article: &pb.Article{
			Id:       "",
			AuthorId: "Bob",
			Title:    "Book1",
			Content:  "Once upon a time",
		},
	}

	s := NewBlogServer(repo.NewMapRepo(make(map[primitive.ObjectID]models.Article)))

	res, err := s.Create(context.Background(), r)

	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	if res.Article.Id == "" {
		t.Fatal("Id should not be empty")
	}
	if res.Article.Title != "Book1" {
		t.Fatalf("Expected Title to be Book1, got %v", res.Article.Title)
	}
}

type mapRepoWithCreateError struct {
	repo.MapArticleRepo
}

func (r *mapRepoWithCreateError) AddArticle(context.Context, *models.Article) (string, error) {
	return "", errors.New("Random error")
}

func TestCreate_repo_error(t *testing.T) {
	r := &pb.CreateRequest{
		Article: &pb.Article{},
	}

	s := BlogServer{
		r: &mapRepoWithCreateError{},
	}
	res, err := s.Create(context.Background(), r)

	if res != nil {
		t.Error("Expect res to be nil")
	}
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Internal {
		t.Errorf("Error status code is not %v, it's %v", codes.Internal.String(), se.Code().String())
	} else if se.Message() != internalError {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
}

func TestCreate_context_cancelled(t *testing.T) {
	r := &pb.CreateRequest{
		Article: &pb.Article{},
	}

	s := BlogServer{
		r: repo.NewMapRepo(make(map[primitive.ObjectID]models.Article)),
	}

	ctx, cancel := context.WithCancel(context.Background())
	// call cancel now
	cancel()

	res, err := s.Create(ctx, r)
	if res != nil {
		t.Error("Expect res to be nil")
	}
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Canceled {
		t.Errorf("Error status code is not %v, it's %v", codes.Canceled.String(), se.Code().String())
	} else if se.Message() != requestCancelled {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
}

func TestRead_success(t *testing.T) {
	h := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	r := &pb.ReadRequest{
		Id: h,
	}

	m := make(map[primitive.ObjectID]models.Article)
	oid, _ := primitive.ObjectIDFromHex(h)
	m[oid] = models.Article{
		ID:    oid,
		Title: "Book2",
	}

	s := BlogServer{
		r: repo.NewMapRepo(m),
	}

	res, err := s.Read(context.Background(), r)
	if err != nil {
		t.Fatalf("Got eror: %v", err)
	}
	if res.Article.Title != "Book2" {
		t.Fatalf("Got wrong article: %v", res.Article)
	}
}

type mapRepoWithReadError struct {
	repo.MapArticleRepo
}

func (r *mapRepoWithReadError) GetArticle(context.Context, string) (*models.Article, error) {
	return nil, errors.New("Read error")
}

func TestRead_error(t *testing.T) {
	h := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	r := &pb.ReadRequest{
		Id: h,
	}

	s := BlogServer{
		r: &mapRepoWithReadError{},
	}

	res, err := s.Read(context.Background(), r)
	if res != nil {
		t.Fatalf("Got result: %v", res)
	}
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Internal {
		t.Errorf("Error status code is not %v, it's %v", codes.Internal.String(), se.Code().String())
	} else if se.Message() != internalError {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
}

func TestRead_context_cancelled(t *testing.T) {
	h := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	r := &pb.ReadRequest{
		Id: h,
	}

	m := make(map[primitive.ObjectID]models.Article)
	oid, _ := primitive.ObjectIDFromHex(h)
	m[oid] = models.Article{
		ID:    oid,
		Title: "Book2",
	}

	s := BlogServer{
		r: repo.NewMapRepo(m),
	}

	ctx, cancel := context.WithCancel(context.Background())
	// call cancel now
	cancel()

	res, err := s.Read(ctx, r)
	if res != nil {
		t.Error("Expect res to be nil")
	}
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Canceled {
		t.Errorf("Error status code is not %v, it's %v", codes.Canceled.String(), se.Code().String())
	} else if se.Message() != requestCancelled {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
}

func TestUpdate_success(t *testing.T) {
	h := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	r := &pb.UpdateRequest{
		Article: &pb.Article{
			Id:    h,
			Title: "Book2_updated",
		},
	}

	m := make(map[primitive.ObjectID]models.Article)
	oid, _ := primitive.ObjectIDFromHex(h)
	m[oid] = models.Article{
		ID:    oid,
		Title: "Book2",
	}

	s := BlogServer{
		r: repo.NewMapRepo(m),
	}

	res, err := s.Update(context.Background(), r)
	if err != nil {
		t.Fatalf("Got eror: %v", err)
	}
	if res.Article.Title != "Book2_updated" {
		t.Fatalf("Got wrong article: %v", *res.Article)
	}
}

type mapRepoWithUpdateError struct {
	repo.MapArticleRepo
}

func (r *mapRepoWithUpdateError) UpdateArticle(context.Context, *models.Article) (*models.Article, error) {
	return nil, errors.New("Update error")
}

func TestUpdate_error_empty_article(t *testing.T) {
	r := &pb.UpdateRequest{
		Article: &pb.Article{},
	}

	s := NewBlogServer(repo.NewMapRepo(make(map[primitive.ObjectID]models.Article)))

	res, err := s.Update(context.Background(), r)
	if res != nil {
		t.Fatalf("Got result: %v", res)
	}
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Internal {
		t.Errorf("Error status code is not %v, it's %v", codes.Internal.String(), se.Code().String())
	} else if se.Message() != internalError {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
}

func TestUpdate_error_repo(t *testing.T) {
	h := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	r := &pb.UpdateRequest{
		Article: &pb.Article{
			Id:    h,
			Title: "Book2_updated",
		},
	}

	s := BlogServer{
		r: &mapRepoWithUpdateError{},
	}

	res, err := s.Update(context.Background(), r)
	if res != nil {
		t.Fatalf("Got result: %v", res)
	}
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Internal {
		t.Errorf("Error status code is not %v, it's %v", codes.Internal.String(), se.Code().String())
	} else if se.Message() != internalError {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
}

func TestUpdate_context_cancelled(t *testing.T) {
	h := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	r := &pb.UpdateRequest{
		Article: &pb.Article{
			Id:    h,
			Title: "Book2_updated",
		},
	}

	m := make(map[primitive.ObjectID]models.Article)
	oid, _ := primitive.ObjectIDFromHex(h)
	m[oid] = models.Article{
		ID:    oid,
		Title: "Book2",
	}

	s := BlogServer{
		r: repo.NewMapRepo(m),
	}

	ctx, cancel := context.WithCancel(context.Background())
	// call cancel now
	cancel()

	res, err := s.Update(ctx, r)
	if res != nil {
		t.Error("Expect res to be nil")
	}
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Canceled {
		t.Errorf("Error status code is not %v, it's %v", codes.Canceled.String(), se.Code().String())
	} else if se.Message() != requestCancelled {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
}

func TestDelete_success(t *testing.T) {
	h := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	r := &pb.DeleteRequest{
		Id: h,
	}

	m := make(map[primitive.ObjectID]models.Article)
	oid, _ := primitive.ObjectIDFromHex(h)
	m[oid] = models.Article{
		ID:    oid,
		Title: "Book2",
	}

	s := BlogServer{
		r: repo.NewMapRepo(m),
	}

	res, err := s.Delete(context.Background(), r)
	if err != nil {
		t.Fatalf("Got eror: %v", err)
	}
	if res.Id != h {
		t.Fatalf("Got wrong article: %v", res.Id)
	}
	if len(m) > 0 {
		t.Fatal("Article was not deleted")
	}
}

type mapRepoWithDeleteError struct {
	repo.MapArticleRepo
}

func (r *mapRepoWithDeleteError) DeleteArticle(context.Context, string) (*models.Article, error) {
	return nil, errors.New("Delete error")
}

func TestDelete_error(t *testing.T) {
	h := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	r := &pb.DeleteRequest{
		Id: h,
	}

	s := BlogServer{
		r: &mapRepoWithReadError{},
	}

	res, err := s.Delete(context.Background(), r)
	if res != nil {
		t.Fatalf("Got result: %v", res)
	}
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Internal {
		t.Errorf("Error status code is not %v, it's %v", codes.Internal.String(), se.Code().String())
	} else if se.Message() != internalError {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
}

func TestDelete_context_cancelled(t *testing.T) {
	h := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	r := &pb.DeleteRequest{
		Id: h,
	}

	m := make(map[primitive.ObjectID]models.Article)
	oid, _ := primitive.ObjectIDFromHex(h)
	m[oid] = models.Article{
		ID:    oid,
		Title: "Book2",
	}

	s := BlogServer{
		r: repo.NewMapRepo(m),
	}

	ctx, cancel := context.WithCancel(context.Background())
	// call cancel now
	cancel()

	res, err := s.Delete(ctx, r)
	if res != nil {
		t.Error("Expect res to be nil")
	}
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Canceled {
		t.Errorf("Error status code is not %v, it's %v", codes.Canceled.String(), se.Code().String())
	} else if se.Message() != requestCancelled {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
}

type testServer struct {
	grpc.ServerStream

	articles []*pb.Article
}

func (s *testServer) Send(m *pb.ListResponse) error {
	s.articles = append(s.articles, m.Article)
	return nil
}

func TestList_all(t *testing.T) {
	h1 := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	h2 := hex.EncodeToString([]byte{12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1})
	m := make(map[primitive.ObjectID]models.Article, 2)
	oid1, _ := primitive.ObjectIDFromHex(h1)
	oid2, _ := primitive.ObjectIDFromHex(h2)
	m[oid1] = models.Article{
		ID:    oid1,
		Title: "Book11",
	}
	m[oid2] = models.Article{
		ID:    oid2,
		Title: "Book12",
	}

	r := &pb.ListRequest{}

	s := BlogServer{
		r: repo.NewMapRepo(m),
	}

	ts := &testServer{articles: []*pb.Article{}}

	err := s.List(r, ts)
	if err != nil {
		t.Fatalf("Got error back: %v", err)
	}
	if len(ts.articles) != 2 {
		t.Fatalf("Wrong number of values in slice. Expected 2, slice: %v", ts.articles)
	}
}

type sleepyServer struct {
	grpc.ServerStream

	articles []*pb.Article
}

func (s *sleepyServer) Send(m *pb.ListResponse) error {
	s.articles = append(s.articles, m.Article)
	time.Sleep(time.Duration(50 * time.Millisecond))
	return nil
}

func TestList_context_expired(t *testing.T) {
	h1 := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	h2 := hex.EncodeToString([]byte{12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1})
	m := make(map[primitive.ObjectID]models.Article, 2)
	oid1, _ := primitive.ObjectIDFromHex(h1)
	oid2, _ := primitive.ObjectIDFromHex(h2)
	m[oid1] = models.Article{
		ID:    oid1,
		Title: "Book11",
	}
	m[oid2] = models.Article{
		ID:    oid2,
		Title: "Book12",
	}

	r := &pb.ListRequest{}

	s := BlogServer{
		r: repo.NewMapRepo(m),
	}

	ts := &sleepyServer{articles: []*pb.Article{}}

	// set timeout to 1 millisecond
	ListTimeout = time.Duration(40 * time.Millisecond)
	err := s.List(r, ts)
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.DeadlineExceeded {
		t.Errorf("Error status code is not %v, it's %v", codes.DeadlineExceeded.String(), se.Code().String())
	} else if se.Message() != internalError {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
	if len(ts.articles) != 1 {
		t.Fatalf("Wrong number of values in slice. Expected 1, slice: %v", ts.articles)
	}
}

type sloppyServer struct {
	grpc.ServerStream

	articles []*pb.Article
}

func (s *sloppyServer) Send(m *pb.ListResponse) error {
	return errors.New("oops I did it again")
}

func TestList_error_sending(t *testing.T) {
	h1 := hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	h2 := hex.EncodeToString([]byte{12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1})
	m := make(map[primitive.ObjectID]models.Article, 2)
	oid1, _ := primitive.ObjectIDFromHex(h1)
	oid2, _ := primitive.ObjectIDFromHex(h2)
	m[oid1] = models.Article{
		ID:    oid1,
		Title: "Book11",
	}
	m[oid2] = models.Article{
		ID:    oid2,
		Title: "Book12",
	}

	r := &pb.ListRequest{}

	s := BlogServer{
		r: repo.NewMapRepo(m),
	}

	ts := &sloppyServer{articles: []*pb.Article{}}

	err := s.List(r, ts)
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Internal {
		t.Errorf("Error status code is not %v, it's %v", codes.Internal.String(), se.Code().String())
	} else if se.Message() != internalError {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
	if len(ts.articles) > 0 {
		t.Fatalf("Wrong number of values in slice. Expected 0, slice: %v", ts.articles)
	}
}

type mapRepoWithFillError struct {
	repo.MapArticleRepo
}

func (r *mapRepoWithFillError) FillArticles(context.Context, chan<- models.Article, <-chan struct{}) error {
	return errors.New("Fill error")
}

func TestList_error_filling(t *testing.T) {
	r := &pb.ListRequest{}

	s := BlogServer{
		r: &mapRepoWithFillError{},
	}

	ts := &testServer{articles: []*pb.Article{}}

	err := s.List(r, ts)
	if se, ok := status.FromError(err); !ok {
		t.Error("Could not initialize status from error")
	} else if se.Code() != codes.Internal {
		t.Errorf("Error status code is not %v, it's %v", codes.Internal.String(), se.Code().String())
	} else if se.Message() != internalError {
		t.Errorf("Wrong message: \"%v\"", se.Message())
	}
	if len(ts.articles) > 0 {
		t.Fatalf("Wrong number of values in slice. Expected 0, slice: %v", ts.articles)
	}
}
