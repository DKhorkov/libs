package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Connector interface is created for usage in external application according to
// "dependency inversion principle" of SOLID due to working via abstractions.
//
//go:generate mockgen -source=interfaces.go -destination=mocks/connector.go -package=mocks -exclude_interfaces=
type Connector interface {
	Database(name string) (*mongo.Database, error)
	Collection(
		database *mongo.Database,
		name string,
		opts ...*mongoOptions.CollectionOptions,
	) (*mongo.Collection, error)
	Ping(ctx context.Context, rp *readpref.ReadPref) error
	ListDatabases(
		ctx context.Context,
		filter bson.D,
		opts ...*mongoOptions.ListDatabasesOptions,
	) (*mongo.ListDatabasesResult, error)
	ListDatabaseNames(
		ctx context.Context,
		filter bson.D,
		opts ...*mongoOptions.ListDatabasesOptions,
	) ([]string, error)
	Close(ctx context.Context) error
}
