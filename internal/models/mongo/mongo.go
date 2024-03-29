package mongo

import (
	"context"

	"github.com/handsomefox/gowarp/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AccountModel struct {
	collection *mongo.Collection
}

func NewAccountModel(ctx context.Context, uri, database, collection string) (*AccountModel, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, models.ErrConnectionFailed
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, models.ErrPingFailed
	}
	coll := client.Database(database).Collection(collection)

	return &AccountModel{collection: coll}, nil
}

func (am *AccountModel) Insert(ctx context.Context, acc *models.Account) (id any, err error) {
	res, err := am.collection.InsertOne(ctx, acc)
	if err != nil {
		return nil, models.ErrInsertFailed
	}

	return res.InsertedID, nil
}

func (am *AccountModel) GetAny(ctx context.Context) (*models.Account, error) {
	acc := &models.Account{}
	res := am.collection.FindOne(ctx, bson.D{{}})
	if err := res.Decode(acc); err != nil {
		return nil, models.ErrNoRecord
	}

	return acc, nil
}

func (am *AccountModel) Delete(ctx context.Context, id any) error {
	_, err := am.collection.DeleteOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}})
	if err != nil {
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
