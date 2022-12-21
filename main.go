package docdb_poc

import (
	"context"
	"github.com/sirupsen/logrus"
	dbutil "gitscm.cisco.com/mcmp/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type client struct {
	dbc *mongo.Database
}

func NewClient() (*client, error) {
	ctx := context.Background()

	dbc, err := dbutil.New(ctx, true)
	if err != nil {
		return nil, err
	}

	return &client{dbc: dbc}, nil
}

func main() {
	// connect with database
	database, err := NewClient()
	if err != nil {
		panic(err)
	}

	collection := database.dbc.Collection("collection")

	// save data
	res, err := collection.InsertOne(context.TODO(), bson.M{"name": "Runon MCMP"})
	if err != nil {
		panic(err)
	}

	logrus.Infof("Inserted document with ID:%v", res.InsertedID)
}
