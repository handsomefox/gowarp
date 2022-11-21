package mongo

import (
	"context"
	"log"

	"github.com/handsomefox/gowarp/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AccountModel struct {
	collection *mongo.Collection
}

func NewAccountModel(ctx context.Context, uri string) (*AccountModel, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Println(err)
		return nil, models.ErrConnectionFailed
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Println(err)
		return nil, models.ErrPingFailed
	}
	collection := client.Database("gowarp").Collection("keys")

	log.Println("connected to the database")

	return &AccountModel{collection: collection}, nil
}

func (am *AccountModel) Insert(ctx context.Context, acc *models.Account) (id any, err error) {
	res, err := am.collection.InsertOne(ctx, acc)
	if err != nil {
		log.Println(err)
		return nil, models.ErrInsertFailed
	}

	return res.InsertedID, nil
}

func (am *AccountModel) GetAny(ctx context.Context) (*models.Account, error) {
	acc := &models.Account{}
	res := am.collection.FindOne(ctx, bson.D{{}})
	if err := res.Decode(acc); err != nil {
		log.Println(err)
		return nil, models.ErrNoRecord
	}

	return acc, nil
}

func (am *AccountModel) Delete(ctx context.Context, id any) error {
	_, err := am.collection.DeleteOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}})
	if err != nil {
		log.Println(err)
		return models.ErrDeleteFailed
	}

	return nil
}

func (am *AccountModel) Len(ctx context.Context) int64 {
	i, err := am.collection.CountDocuments(ctx, bson.D{{}})
	if err != nil {
		return 0
	}

	return i
}
