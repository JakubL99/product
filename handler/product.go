package handler

import (
	"context"
	"fmt"

	pb "product/proto"

	"github.com/micro/micro/v3/service/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type rProduct struct {
	ID           primitive.ObjectID `bson:"_id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Platform     string             `json:"platform"`
	Version      string             `json:"version"`
	Pegi         string             `json:"pegi"`
	Price        string             `json:"price"`
	Availability int32              `json:"availability"`
}

type Product struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Platform     string `json:"platform"`
	Version      string `json:"version"`
	Pegi         string `json:"pegi"`
	Price        string `json:"price"`
	Availability int32  `json:"availability"`
}

type idFilter struct {
	ID primitive.ObjectID `bson:"_id"`
}

type Handler struct {
	Repo
}

type Repo interface {
	Create(ctx context.Context, product *Product) error
	GetAll(ctx context.Context) ([]*rProduct, error)
	FindOne(ctx context.Context, id *idFilter) (*rProduct, error)
	DeleteOne(ctx context.Context, id *idFilter) (*rProduct, error)
}

type MongoRepository struct {
	Collection *mongo.Collection
}

//Conversion function Json <-> Protobuf
func MarshalProduct(product *pb.CreateRequest) *Product {
	return &Product{
		Name:         product.Name,
		Description:  product.Description,
		Platform:     product.Platform,
		Version:      product.Version,
		Pegi:         product.Pegi,
		Price:        product.Price,
		Availability: product.Availability,
	}
}

func MarshalID(id *pb.GetRequest) *idFilter {
	oid, err := primitive.ObjectIDFromHex(id.Id)
	if err != nil {
		logger.Error("Error conversing string to ObjectIDFromHex ", err)
	}
	return &idFilter{
		ID: oid,
	}
}

func MarshalIDD(id *pb.DeleteRequest) *idFilter {
	oid, err := primitive.ObjectIDFromHex(id.Id)
	if err != nil {
		logger.Error("Error conversing string to ObjectIDFromHex ", err)
	}
	return &idFilter{
		ID: oid,
	}
}

func UnmarshalProduct(product *rProduct) *pb.CreateRequest {
	obj_id := primitive.ObjectID.Hex(product.ID)
	return &pb.CreateRequest{
		Id:           obj_id,
		Name:         product.Name,
		Description:  product.Description,
		Platform:     product.Platform,
		Version:      product.Version,
		Pegi:         product.Pegi,
		Price:        product.Price,
		Availability: product.Availability,
	}
}

func UnmarshalProductCollection(products []*rProduct) []*pb.CreateRequest {
	collection := make([]*pb.CreateRequest, 0)
	for _, product := range products {
		collection = append(collection, UnmarshalProduct(product))
	}
	return collection
}

//A function that operates on a database
func (repo *MongoRepository) Create(ctx context.Context, product *Product) error {
	_, err := repo.Collection.InsertOne(ctx, product)
	return err
}

func (repo *MongoRepository) GetAll(ctx context.Context) ([]*rProduct, error) {
	cur, err := repo.Collection.Find(ctx, bson.M{}, nil)
	if err != nil {
		logger.Error("Error reading ", err)
	}
	defer cur.Close(ctx)
	var products []*rProduct
	for cur.Next(ctx) {
		var product *rProduct
		if err = cur.Decode(&product); err != nil {
			logger.Error("Error Decode ", err)
			return nil, err
		}
		products = append(products, product)
	}
	return products, err
}

func (repo *MongoRepository) FindOne(ctx context.Context, id *idFilter) (*rProduct, error) {
	cur := repo.Collection.FindOne(ctx, bson.M{"_id": id.ID})
	var product *rProduct
	if err := cur.Decode(&product); err != nil {
		logger.Error("Error Decode ", err)
		return nil, err
	}

	return product, cur.Err()
}

func (repo *MongoRepository) DeleteOne(ctx context.Context, id *idFilter) (*rProduct, error) {
	cur := repo.Collection.FindOne(ctx, bson.M{"_id": id.ID})
	var product *rProduct
	if err := cur.Decode(&product); err != nil {
		logger.Error("Error Decode ", err)
		return nil, err
	}
	resultDelete, err := repo.Collection.DeleteOne(ctx, bson.M{"_id": id.ID})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	logger.Info("resultDelete: ", resultDelete.DeletedCount)
	return product, cur.Err()
}

//Endpoint API
func (h *Handler) Call(ctx context.Context, req *pb.CallRequest, rsp *pb.CallResponse) error {
	logger.Info("Received Product.Call request")
	rsp.Call = "Hello " + req.Name + " in Product Service"
	return nil
}

func (h *Handler) CreateProduct(ctx context.Context, req *pb.CreateRequest, rsp *pb.CreateResponse) error {
	if err := h.Create(ctx, MarshalProduct(req)); err != nil {
		return err
	}
	rsp.Created = true
	rsp.Product = req
	return nil
}

func (h *Handler) List(ctx context.Context, req *pb.ListRequest, rsp *pb.ListResponse) error {
	products, err := h.GetAll(ctx)
	if err != nil {
		logger.Error("Error GetAll", err)
		return err
	}
	rsp.Products = UnmarshalProductCollection(products)
	return nil
}

func (h *Handler) GetProduct(ctx context.Context, req *pb.GetRequest, rsp *pb.GetResponse) error {
	id := MarshalID(req)
	product, err := h.FindOne(ctx, id)
	if err != nil {

	}
	unp := UnmarshalProduct(product)
	rsp.Product = unp
	return nil
}

func (h *Handler) DeleteProduct(ctx context.Context, req *pb.DeleteRequest, rsp *pb.DeleteResponse) error {
	id := MarshalIDD(req)
	product, err := h.DeleteOne(ctx, id)
	if err != nil {
		logger.Error("Error method DeleteOne: ", err)
	}
	unp := UnmarshalProduct(product)
	rsp.Deleted = true
	rsp.Product = unp
	return nil
}
