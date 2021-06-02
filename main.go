package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	handler "product/handler"
	pb "product/proto"

	"github.com/micro/micro/v3/service"
	"github.com/micro/micro/v3/service/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateClient(ctx context.Context, uri string, retry int32) (*mongo.Client, error) {
	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err := conn.Ping(ctx, nil); err != nil {
		if retry >= 3 {
			fmt.Printf("Failed connect")
			return nil, err
		}
		retry = retry + 1
		time.Sleep(time.Second * 2)
		return CreateClient(ctx, uri, retry)
	}
	fmt.Printf("Connect with database")
	return conn, err
}

func main() {
	uri := os.Getenv("DB_HOST")

	srv := service.New(
		service.Name("product"),
		service.Version("latest"),
	)

	//srv.Init()

	client, err := CreateClient(context.Background(), uri, 0)
	if err != nil {
		log.Panic(err)
	}
	defer client.Disconnect(context.Background())

	productCollection := client.Database("product").Collection("games")

	repo := &handler.MongoRepository{
		Collection: productCollection,
	}

	h := &handler.Handler{
		Repo: repo,
	}

	pb.RegisterProductHandler(srv.Server(), h)

	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
