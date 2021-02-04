package models

import (
	pb "example.com/grpc/blog/gen/src"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Article model represents the article
type Article struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
}

// FromPB creates Article from Protocol Buffers struct definition
func FromPB(a *pb.Article) (*Article, error) {
	oid, err := primitive.ObjectIDFromHex(a.GetId())
	if err != nil {
		return nil, err
	}

	return &Article{
		ID:       oid,
		AuthorID: a.GetAuthorId(),
		Title:    a.GetTitle(),
		Content:  a.GetContent(),
	}, nil
}

// ToPB converts Article to Protocol Buffer message
func (a Article) ToPB() *pb.Article {
	return &pb.Article{
		Id:       a.ID.String(),
		AuthorId: a.AuthorID,
		Title:    a.Title,
		Content:  a.Content,
	}
}
