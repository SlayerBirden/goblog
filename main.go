package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "example.com/grpc/blog/gen/src"
	"example.com/grpc/blog/src/repo"
	"example.com/grpc/blog/src/server"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	c, err := initDB()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2*time.Second))
		defer cancel()
		if err = c.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	if err != nil {
		log.Fatalln("Error getting Mongo client", err)
	}
	r := repo.NewMongoArticleRepo(c)
	log.Fatal(serve(r))
}

func initDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2*time.Second))
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func serve(r repo.ArticleRepo) error {
	li, err := net.Listen("tcp", os.Getenv("URI"))
	if err != nil {
		return err
	}
	defer li.Close()
	s := grpc.NewServer()
	pb.RegisterBlogServer(s, server.NewBlogServer(r))
	reflection.Register(s)
	fmt.Println("Listening on", os.Getenv("URI"), "...")
	if err := s.Serve(li); err != nil {
		return err
	}
	return nil
}
