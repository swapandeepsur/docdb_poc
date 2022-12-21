/*
Package mongo provides common functionality for using the official MongoDB driver (SDK).

Constructor for creation of a configured MongoDB database connection using the standard configurations used within across MCMP.

Example Constructor usage within a service (read-write)

	type client struct {
		dbc *mongo.Database
	}

	func NewClient(opts *db.Options) (db.Datastore, error) {
		ctx := context.Background()

		if avail.IsFlagEnabled(ctx, config.FeatureEnableMongoFuture) {
			dbcfg.Future()
		}

		dbc, err := dbutil.New(ctx)
		if err != nil {
			return nil, err
		}

		if opts.TestMode {
			seedDB(ctx, dbc)
		}

		return &client{dbc: dbc}, nil
	}

Example Constructor usage within a service (read-only)

	type client struct {
		dbc *mongo.Database
	}

	func NewClient(opts *db.Options) (db.Datastore, error) {
		ctx := context.Background()

		if avail.IsFlagEnabled(ctx, config.FeatureEnableMongoFuture) {
			dbcfg.Future()
		}

		dbc, err := dbutil.New(ctx, dbutil.ReadOnly())
		if err != nil {
			return nil, err
		}

		if opts.TestMode {
			seedDB(ctx, dbc)
		}

		return &client{dbc: dbc}, nil
	}

Converting a search query into a MongoDB query.

Example Conversion usage for a simple List.

	type client struct {
		dbc *mongo.Database
	}

	func (c *client) ListGroups(ctx context.Context, query *search.Query) ([]*domain.Group, error) {
		cur, err := c.dbc.Collection(groupsCollection).Find(ctx, dbutil.Filters(query), dbutil.FindOptions(query))
		if err != nil {
			return nil, err
		}

		var groups []*domain.Group

		if err := cur.All(ctx, &groups); err != nil {
			return nil, err
		}

		return groups, nil
	}

Example Conversion usage with pagination.

	type client struct {
		dbc *mongo.Database
	}

	func (c *client) ListGroups(ctx context.Context, query *search.Query) ([]*domain.Group, error) {
		// execute the query to get total count if the results are sorted
		if !query.EmptySortby() {
			query.Count, _ = c.dbc.Collection(groupsCollection).CountDocuments(ctx, dbutil.Filters(query))
		}

		cur, err := c.dbc.Collection(groupsCollection).Find(ctx, dbutil.Filters(query), dbutil.FindOptions(query))
		if err != nil {
			return nil, err
		}

		var groups []*domain.Group

		if err := cur.All(ctx, &groups); err != nil {
			return nil, err
		}

		return groups, nil
	}

Example index model configuration and management.

	var indexModels = index.NewModels(
		// Barebones index using default settings
		index.New(
			"index_name", bson.D{{Key: "fieldName", Value: 1}},
		),

		// Partial index using PartialFilterExpression
		index.New(
			"index_name", bson.D{{Key: "fieldName", Value: 1}}, index.PartialFilter(bson.M{"field": "value"}),
		),

		// TTL index using ExpireAfterSeconds
		index.New(
			"index_name", bson.D{{Key: "fieldName", Value: 1}}, index.ExpireAfter(3600),
		),

		// Sparse TTL index (combining options)
		index.New(
			"index_name", bson.D{{Key: "fieldName", Value: 1}}, index.Sparse(true), index.ExpireAfter(3600),
		),

		// Index with custom options and validator
		index.New(
			"index_name", bson.D{{Key: "fieldName", Value: 1}},
			index.Option(options.Index().SetCollation(&options.Collation{Backwards: true}), customValidateFunc),
		),
	)

	func customValidateFunc(index bson.M, model mongo.IndexModel) bool {
		col, ok := index["collation"]
		if !ok {
			return false
		}

		return reflect.DeepEqual(model.Options.Collation.ToDocument(), col)
	}

	func (c *client) ApplyIndexes(ctx context.Context, collectionName string) error {
		// ensure that all policy indexes exist and are up to date
		return indexModels.Apply(ctx, c.dbc.Collection(collectionName))
	}
*/
package mongo
