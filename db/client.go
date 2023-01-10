package db

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type Datastore interface {
	SaveData(ctx context.Context, object bson.M) error
}

// DatastoreFactory is a type for data store factory methods.
type DatastoreFactory func() (Datastore, error)

var datastoreFactories = make(map[string]DatastoreFactory)

// Register adds a DatastoreFactory for usage.
func Register(name string, factory DatastoreFactory) {
	if factory == nil {
		logrus.Panicf("Datastore factory %s did not provide initialization function.", name)
	}

	_, registered := datastoreFactories[name]
	if registered {
		logrus.Warnf("Datastore factory %s already registered. Ignoring.", name)

		return
	}

	datastoreFactories[name] = factory
}
