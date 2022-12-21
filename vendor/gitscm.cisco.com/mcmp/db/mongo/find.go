package mongo

import (
	"regexp"
	"time"

	"gitscm.cisco.com/mcmp/utils/search"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	pk    = "_id"
	idRef = "id"
)

// FindOptions uses the search Query to construct a mongodb FindOptions.
func FindOptions(q *search.Query) *options.FindOptions {
	opts := options.Find()

	if !q.EmptyFields() {
		opts.SetProjection(Select(q))
	}

	if !q.EmptySortby() {
		opts.SetSort(Sort(q))
	}

	if q.Limit() > 0 {
		opts.SetLimit(int64(q.Limit()))
	}

	if !q.EmptySortby() && q.Offset() > 0 {
		opts.SetSkip(int64(q.Offset()))
	}

	return opts
}

// Select uses the search Query to construct a mongodb Projection input.
func Select(q *search.Query) bson.D {
	selector := make(bson.D, 0)

	for _, f := range q.Fields() {
		selector = append(selector, bson.E{Key: id(f), Value: 1})
	}

	return selector
}

// Sort uses the search Query to construct a mongodb Sort input.
func Sort(q *search.Query) bson.D {
	sb := make(bson.D, 0)

	iter := q.Sortby().EntriesIter()

	for {
		pair, ok := iter()
		if !ok {
			break
		}

		if pair.Value.(bool) {
			sb = append(sb, bson.E{Key: id(pair.Key), Value: 1})
		} else {
			sb = append(sb, bson.E{Key: id(pair.Key), Value: -1})
		}
	}

	return sb
}

// Filters uses the search Query to construct a mongodb Collection.Find input.
func Filters(q *search.Query) bson.M {
	qf := make(bson.M)

	for k, f := range q.Filters() {
		switch mval := f.Value.(type) {
		case []string:
			qf[id(k)] = bson.M{"$in": mval}
		case time.Time, int, int32, int64, float32, float64:
			qf[id(k)] = asGenericComparison(f, mval)
		case string:
			qf[id(k)] = asStringComparison(f, mval)
		default:
			qf[id(k)] = mval
		}
	}

	return qf
}

func asGenericComparison(f search.Filter, value interface{}) bson.M {
	if f.LTE() {
		return bson.M{"$lte": value}
	}

	if f.GTE() {
		return bson.M{"$gte": value}
	}

	return bson.M{"$eq": value}
}

func asStringComparison(f search.Filter, val string) bson.M {
	if f.Like() {
		return bson.M{"$regex": primitive.Regex{Pattern: "^" + regexp.QuoteMeta(val), Options: "i"}}
	}

	if f.IgnoreCase() {
		return bson.M{"$regex": primitive.Regex{Pattern: "^" + regexp.QuoteMeta(val) + "$", Options: "i"}}
	}

	if f.NotEqual() {
		return bson.M{"$ne": val}
	}

	return bson.M{"$eq": val}
}

func id(attr string) string {
	// handles converting to mongodb naming of the ID attribute
	if attr == idRef {
		return pk
	}

	return attr
}
