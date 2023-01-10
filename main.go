package docdb_poc

import (
	"context"
	"github.com/sirupsen/logrus"
	dbutil "gitscm.cisco.com/mcmp/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	db "docdb_poc/db"
)

func init() {
	db.Register("mongodb", NewClient)
}

type client struct {
	dbc *mongo.Database
}

func NewClient() (db.Datastore, error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "flag", true)

	dbc, err := dbutil.New(ctx)
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

	// save data
	if err = database.SaveData(context.TODO(), bson.M{"name": "Runon MCMP Test"}); err != nil {
		panic(err)
	}
}

func (c *client) SaveData(ctx context.Context, object bson.M) error {
	collection := c.dbc.Collection("collection")

	// insert record
	res, err := collection.InsertOne(context.TODO(), object)
	if err != nil {
		panic(err)
	}

	logrus.Infof("Inserted document with ID:%v", res.InsertedID)

	return nil
}
