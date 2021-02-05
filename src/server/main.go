package server

import (
	"context"
	"log"
	"time"

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
	id, err := s.r.AddArticle(ctx, m)
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

// Read returns one Article Doc
func (s *BlogServer) Read(ctx context.Context, r *pb.ReadRequest) (*pb.ReadResponse, error) {
	m, err := s.r.GetArticle(ctx, r.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.ReadResponse{Article: m.ToPB()}, nil
}

// List streams Articles
func (s *BlogServer) List(r *pb.ListRequest, stream pb.Blog_ListServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(100*time.Second))
	defer cancel()
	interrupt := make(chan struct{})
	out := make(chan models.Article)
	e := make(chan error)
	go func() {
		defer close(e)
		for v := range out {
			if err := stream.Send(&pb.ListResponse{Article: v.ToPB()}); err != nil {
				// dump last value from out
				// Otherwise we're stuck because message was already sent and waiting for next iteration to read it
				// before "interrupt" signal can be read
				<-out
				interrupt <- struct{}{}
				e <- status.Errorf(codes.Internal, "Got error while sending: %v", err)
				return
			}
		}
		e <- nil
	}()
	err := s.r.FillArticles(ctx, out, interrupt)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return <-e
}

// Update an Article and returns result with updated Article
func (s *BlogServer) Update(ctx context.Context, r *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	m, err := models.FromPB(r.GetArticle())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error updating article: %v", err)
	}
	res, err := s.r.UpdateArticle(ctx, m)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error updating article: %v", err)
	}
	return &pb.UpdateResponse{Article: res.ToPB()}, status.Error(codes.OK, "Successfully updated Article")
}

// Delete an Article by ID
func (s *BlogServer) Delete(ctx context.Context, r *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	m, err := s.r.DeleteArticle(ctx, r.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error deleting article: %v", m)
	}
	return &pb.DeleteResponse{Id: m.ID.Hex()}, status.Error(codes.OK, "Successfully deleted Article")
}
