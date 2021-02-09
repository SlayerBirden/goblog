package server

import (
	"context"
	"fmt"
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
	internalError    = "There was an error internally"
	requestCancelled = "Client cancelled request, aborting"
)

// ListTimeout controls how much time List waits until cancelling
var ListTimeout time.Duration = time.Duration(5 * time.Second)

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
		return nil, status.Error(codes.Canceled, requestCancelled)
	}
	a.Id = id
	return &pb.CreateResponse{Article: a}, status.Error(codes.OK, "Successfully created the article")
}

// Read returns one Article Doc
func (s *BlogServer) Read(ctx context.Context, r *pb.ReadRequest) (*pb.ReadResponse, error) {
	m, err := s.r.GetArticle(ctx, r.GetId())
	if err != nil {
		log.Printf("Error while reading: %v\n", err)
		return nil, status.Error(codes.Internal, internalError)
	}
	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, requestCancelled)
	}
	return &pb.ReadResponse{Article: m.ToPB()}, nil
}

// List streams Articles
func (s *BlogServer) List(r *pb.ListRequest, stream pb.Blog_ListServer) error {
	ctx, cancel := context.WithTimeout(context.Background(), ListTimeout)
	defer cancel()
	stop := make(chan struct{})
	out := make(chan models.Article)
	e := make(chan error)
	go func() {
		defer close(e)
		for v := range out {
			if ctx.Err() == context.DeadlineExceeded {
				interruptList("Exceeded deadline", out, stop, e, codes.DeadlineExceeded)
				return
			}
			if err := stream.Send(&pb.ListResponse{Article: v.ToPB()}); err != nil {
				interruptList(fmt.Sprintf("Got error while sending: %v", err), out, stop, e, codes.Internal)
				return
			}
		}
		// send signal to close "stop" goroutine
		stop <- struct{}{}
		e <- nil
	}()
	err := s.r.FillArticles(ctx, out, stop)
	if err != nil {
		log.Printf("Error when filling out channel: %v", err)
		return status.Error(codes.Internal, internalError)
	}

	return <-e
}

// inerruptList used to interrup Fetching of new Articles
func interruptList(msg string, out <-chan models.Article, stop chan<- struct{}, e chan<- error, code codes.Code) {
	// dump last value from out
	// Otherwise we're stuck because message was already sent and waiting for next iteration to read it
	// before "interrupt" signal can be read
	<-out
	stop <- struct{}{}
	log.Print(msg)
	e <- status.Errorf(code, internalError)
}

// Update an Article and returns result with updated Article
func (s *BlogServer) Update(ctx context.Context, r *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	m, err := models.FromPB(r.GetArticle())
	if err != nil {
		log.Printf("Error updating article: %v\n", err)
		return nil, status.Error(codes.Internal, internalError)
	}
	res, err := s.r.UpdateArticle(ctx, m)
	if err != nil {
		log.Printf("Error updating article: %v\n", err)
		return nil, status.Error(codes.Internal, internalError)
	}
	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, requestCancelled)
	}
	return &pb.UpdateResponse{Article: res.ToPB()}, status.Error(codes.OK, "Successfully updated Article")
}

// Delete an Article by ID
func (s *BlogServer) Delete(ctx context.Context, r *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	m, err := s.r.DeleteArticle(ctx, r.GetId())
	if err != nil {
		log.Printf("Error deleting article: %v\n", err)
		return nil, status.Error(codes.Internal, internalError)
	}
	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, requestCancelled)
	}
	return &pb.DeleteResponse{Id: m.ID.Hex()}, status.Error(codes.OK, "Successfully deleted Article")
}
