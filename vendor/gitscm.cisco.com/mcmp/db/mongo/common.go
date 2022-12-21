package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
)

// PK creates a primary key doc using the provided id.
func PK(id string) bson.D {
	return bson.D{{Key: pk, Value: id}}
}
