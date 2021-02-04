package server

import (
	"context"
	"log"

	pb "example.com/grpc/blog/gen/src"
	"example.com/grpc/blog/src/models"
	"example.com/grpc/blog/src/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	internalError = "There was an error internally"
)

// BlogServer implements GRPC sever for our blog
type BlogServer struct {
	pb.UnimplementedBlogServer

	r repo.ArticleRepo
}

// NewBlogServer returns a blogServer
func NewBlogServer(r repo.ArticleRepo) *BlogServer {
	return &BlogServer{
		r: r,
	}
}

// Create implements the Create method for our Blog
func (s *BlogServer) Create(ctx context.Context, r *pb.CreateRequest) (*pb.CreateResponse, error) {
	a := r.GetArticle()
	// need to create an Article since ID should be skipped
	m := &models.Article{
		ID:       primitive.NilObjectID,
		AuthorID: a.GetAuthorId(),
		Title:    a.GetTitle(),
		Content:  a.GetContent(),
	}
	id, err := s.r.AddArticle(m)
	if err != nil {
		log.Println("Got error from repo.AddArticle", err)
		return nil, status.Error(codes.Internal, internalError)
	}

	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, "Client cancelled request, abandoning.")
	}
	a.Id = id
	return &pb.CreateResponse{Article: a}, status.Error(codes.OK, "Successfully created the article")
}
