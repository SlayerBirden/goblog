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
)

func main() {
	c, err := initDB()
	if err != nil {
		log.Fatalln("Error getting Mongo client", err)
	}
	r := repo.NewMongoArticleRepo(c)
	serve(r)
}

func initDB() (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20*time.Second))
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func serve(r repo.ArticleRepo) {
	li, err := net.Listen("tcp", os.Getenv("URI"))
	if err != nil {
		log.Fatalln("Failed to start listening", err)
	}
	defer li.Close()
	s := grpc.NewServer()
	pb.RegisterBlogServer(s, server.NewBlogServer(r))
	fmt.Println("Listening on", os.Getenv("URI"), "...")
	if err := s.Serve(li); err != nil {
		log.Fatalln("Error running server", err)
	}
}
