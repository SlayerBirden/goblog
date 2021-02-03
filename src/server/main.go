package server

import (
	"context"
	"log"

	pb "example.com/grpc/blog/gen/src"
	"example.com/grpc/blog/src/models"
	"example.com/grpc/blog/src/repo"
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
func (s *BlogServer) Create(ctx context.Context, r *pb.ArticleMessage) (*pb.ArticleMessage, error) {
	a := r.GetArticle()
	m, err := models.FromPB(a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "There was an error retrieving article: %s", err)
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
	return &pb.ArticleMessage{Article: a}, status.Error(codes.OK, "Successfully created the article")
}
