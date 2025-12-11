package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// New is constructor of CommonConnector. Gets database Config and logging.Logger to create an instance.
func New(
	ctx context.Context,
	dsn string,
	rp *readpref.ReadPref,
	opts ...Option,
) (*CommonConnector, error) {
	client, err := connect(ctx, dsn, rp, opts...)
	if err != nil {
		return nil, err
	}

	dbConnector := &CommonConnector{
		client: client,
	}

	return dbConnector, nil
}

// connect connects to database and stores connections pool for later usage.
func connect(ctx context.Context, dsn string, rp *readpref.ReadPref, opts ...Option) (*mongo.Client, error) {
	var connectOpts options
	for _, opt := range opts {
		err := opt(&connectOpts)
		if err != nil {
			return nil, err
		}
	}

	clientOptions := mongoOptions.Client().ApplyURI(dsn)

	if connectOpts.username != "" && connectOpts.password != "" && connectOpts.authSource != "" {
		clientOptions.SetAuth(
			mongoOptions.Credential{
				Username:   connectOpts.username,
				Password:   connectOpts.password,
				AuthSource: connectOpts.authSource,
			},
		)
	}

	clientOptions.SetMaxConnecting(connectOpts.maxConnections)
	clientOptions.SetConnectTimeout(connectOpts.maxConnectionTimeout)
	clientOptions.SetMaxConnIdleTime(connectOpts.maxConnectionIdleTimeout)
	clientOptions.SetMaxPoolSize(connectOpts.maxPoolSize)
	clientOptions.SetMinPoolSize(connectOpts.minPoolSize)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Проверка соединения
	if err = client.Ping(ctx, rp); err != nil {
		return nil, err
	}

	return client, nil
}

// CommonConnector is base connector to work with database.
type CommonConnector struct {
	client *mongo.Client
}

// Database returns mongo.Database.
func (c *CommonConnector) Database(name string) (*mongo.Database, error) {
	if c.client == nil {
		return nil, &NilClientError{}
	}

	return c.client.Database(name), nil
}

// Collection returns mongo.Collection for provided database.
func (c *CommonConnector) Collection(
	database *mongo.Database,
	name string,
	opts ...*mongoOptions.CollectionOptions,
) (*mongo.Collection, error) {
	if database == nil {
		return nil, &NilDatabaseError{}
	}

	return database.Collection(name, opts...), nil
}

// Ping checks if mongo.Client is connected.
func (c *CommonConnector) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	if c.client == nil {
		return &NilClientError{}
	}

	return c.client.Ping(ctx, rp)
}

// ListDatabases returns all mongo.Database within established connection with mongoDB.
func (c *CommonConnector) ListDatabases(
	ctx context.Context,
	filter bson.D,
	opts ...*mongoOptions.ListDatabasesOptions,
) (*mongo.ListDatabasesResult, error) {
	if c.client == nil {
		return nil, &NilClientError{}
	}

	databases, err := c.client.ListDatabases(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	return &databases, nil
}

// ListDatabaseNames returns all mongo.Database names within established connection with mongoDB.
func (c *CommonConnector) ListDatabaseNames(
	ctx context.Context,
	filter bson.D,
	opts ...*mongoOptions.ListDatabasesOptions,
) ([]string, error) {
	if c.client == nil {
		return nil, &NilClientError{}
	}

	return c.client.ListDatabaseNames(ctx, filter, opts...)
}

// Close closes pool of connections.
func (c *CommonConnector) Close(ctx context.Context) error {
	if c.client == nil {
		return nil
	}

	return c.client.Disconnect(ctx)
}
