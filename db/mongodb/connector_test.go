//go:build integration

package mongodb

import (
	"context"
	"github.com/DKhorkov/libs/db/mongodb/mocks"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"testing"
	"time"
)

/*
Поднять БД перед тестами


docker run -d \
  --name mongodb \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=secret \
  -e MONGO_INITDB_DATABASE=admin \
  -v mongodb_data:/data/db \
  mongo:latest
*/

// TestNew тестирует конструктор CommonConnector
func TestNew(t *testing.T) {
	ctx := context.Background()
	rp := readpref.Primary()

	tests := []struct {
		name       string
		dsn        string
		opts       []Option
		setupMocks func(*mocks.MockConnector)
		wantErr    bool
	}{
		{
			name: "success connection with auth options",
			dsn:  "mongodb://localhost:27017",
			opts: []Option{
				WithUsername("admin"),
				WithPassword("secret"),
				WithAuthSource("admin"),
				WithMaxConnections(100),
				WithMaxPoolSize(50),
				WithMinPoolSize(10),
				WithMaxConnectionTimeout(30 * time.Second),
				WithMaxConnectionIdleTime(5 * time.Minute),
			},
			wantErr: false,
		},
		{
			name:    "invalid dsn",
			dsn:     "invalid://dsn",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector, err := New(ctx, tt.dsn, rp, tt.opts...)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Проверяем, что можно закрыть соединение
				require.NoError(t, connector.Close(ctx))
			}
		})
	}
}

// TestOptionFunctions тестирует функциональные опции
func TestOptionFunctions(t *testing.T) {
	tests := []struct {
		name         string
		option       Option
		validateOpts func(*testing.T, *options)
	}{
		{
			name:   "WithUsername",
			option: WithUsername("testuser"),
			validateOpts: func(t *testing.T, opts *options) {
				if opts.username != "testuser" {
					t.Errorf("username = %s, want testuser", opts.username)
				}
			},
		},
		{
			name:   "WithPassword",
			option: WithPassword("testpass"),
			validateOpts: func(t *testing.T, opts *options) {
				if opts.password != "testpass" {
					t.Errorf("password = %s, want testpass", opts.password)
				}
			},
		},
		{
			name:   "WithAuthSource",
			option: WithAuthSource("admin"),
			validateOpts: func(t *testing.T, opts *options) {
				if opts.authSource != "admin" {
					t.Errorf("authSource = %s, want admin", opts.authSource)
				}
			},
		},
		{
			name:   "WithMaxConnections",
			option: WithMaxConnections(100),
			validateOpts: func(t *testing.T, opts *options) {
				if opts.maxConnections != 100 {
					t.Errorf("maxConnections = %d, want 100", opts.maxConnections)
				}
			},
		},
		{
			name:   "WithMaxPoolSize",
			option: WithMaxPoolSize(50),
			validateOpts: func(t *testing.T, opts *options) {
				if opts.maxPoolSize != 50 {
					t.Errorf("maxPoolSize = %d, want 50", opts.maxPoolSize)
				}
			},
		},
		{
			name:   "WithMinPoolSize",
			option: WithMinPoolSize(10),
			validateOpts: func(t *testing.T, opts *options) {
				if opts.minPoolSize != 10 {
					t.Errorf("minPoolSize = %d, want 10", opts.minPoolSize)
				}
			},
		},
		{
			name:   "WithMaxConnectionTimeout",
			option: WithMaxConnectionTimeout(30 * time.Second),
			validateOpts: func(t *testing.T, opts *options) {
				if opts.maxConnectionTimeout != 30*time.Second {
					t.Errorf("maxConnectionTimeout = %v, want 30s", opts.maxConnectionTimeout)
				}
			},
		},
		{
			name:   "WithMaxConnectionIdleTime",
			option: WithMaxConnectionIdleTime(5 * time.Minute),
			validateOpts: func(t *testing.T, opts *options) {
				if opts.maxConnectionIdleTimeout != 5*time.Minute {
					t.Errorf("maxConnectionIdleTimeout = %v, want 5m", opts.maxConnectionIdleTimeout)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &options{}
			err := tt.option(opts)

			if err != nil {
				t.Errorf("option returned error: %v", err)
			}

			tt.validateOpts(t, opts)
		})
	}
}

func TestCollection(t *testing.T) {
	tests := []struct {
		name                  string
		setupConnector        func() *CommonConnector
		setupDB               func(connector Connector) *mongo.Database
		collectionName        string
		wantErr               bool
		expectedNilCollection bool
	}{
		{
			name: "success get database",
			setupConnector: func() *CommonConnector {
				c, err := New(
					context.Background(),
					"mongodb://localhost:27017",
					nil,
					WithUsername("admin"),
					WithPassword("secret"),
					WithAuthSource("admin"),
				)

				require.NoError(t, err)
				return c
			},
			setupDB: func(connector Connector) *mongo.Database {
				c, err := connector.Database("admin")
				require.NoError(t, err)
				return c
			},
			wantErr:               false,
			expectedNilCollection: false,
		},
		{
			name: "nil db error",
			setupConnector: func() *CommonConnector {
				return &CommonConnector{
					client: nil,
				}
			},
			setupDB: func(_ Connector) *mongo.Database {
				return nil
			},
			wantErr:               true,
			expectedNilCollection: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := tt.setupConnector()
			db := tt.setupDB(connector)
			col, err := connector.Collection(db, tt.collectionName)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if !tt.expectedNilCollection {
				require.NotNil(t, col)
			}
		})
	}
}

func TestDatabase(t *testing.T) {
	tests := []struct {
		name           string
		setupConnector func() *CommonConnector
		dbName         string
		wantErr        bool
		expectedNilDB  bool
	}{
		{
			name: "success get database",
			setupConnector: func() *CommonConnector {
				c, err := New(
					context.Background(),
					"mongodb://localhost:27017",
					nil,
					WithUsername("admin"),
					WithPassword("secret"),
					WithAuthSource("admin"),
				)

				require.NoError(t, err)
				return c
			},
			dbName:        "admin",
			wantErr:       false,
			expectedNilDB: false,
		},
		{
			name: "nil client error",
			setupConnector: func() *CommonConnector {
				return &CommonConnector{
					client: nil,
				}
			},
			dbName:        "testdb",
			wantErr:       true,
			expectedNilDB: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := tt.setupConnector()
			db, err := connector.Database(tt.dbName)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if !tt.expectedNilDB {
				require.NotNil(t, db)
			}
		})
	}
}

func TestPing(t *testing.T) {
	tests := []struct {
		name           string
		setupConnector func() *CommonConnector
		wantErr        bool
	}{
		{
			name: "success ping",
			setupConnector: func() *CommonConnector {
				c, err := New(
					context.Background(),
					"mongodb://localhost:27017",
					nil,
					WithUsername("admin"),
					WithPassword("secret"),
					WithAuthSource("admin"),
				)

				require.NoError(t, err)
				return c
			},
			wantErr: false,
		},
		{
			name: "nil client error",
			setupConnector: func() *CommonConnector {
				return &CommonConnector{
					client: nil,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := tt.setupConnector()
			err := connector.Ping(context.Background(), nil)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name           string
		setupConnector func() *CommonConnector
	}{
		{
			name: "success close",
			setupConnector: func() *CommonConnector {
				c, err := New(
					context.Background(),
					"mongodb://localhost:27017",
					nil,
					WithUsername("admin"),
					WithPassword("secret"),
					WithAuthSource("admin"),
				)

				require.NoError(t, err)
				return c
			},
		},
		{
			name: "nil client",
			setupConnector: func() *CommonConnector {
				return &CommonConnector{
					client: nil,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := tt.setupConnector()
			require.NoError(t, connector.Close(context.Background()))
		})
	}
}

func TestListDatabases(t *testing.T) {
	tests := []struct {
		name           string
		setupConnector func() *CommonConnector
		filter         bson.D
		wantErr        bool
		expected       *mongo.ListDatabasesResult
	}{
		{
			name: "success list databases",
			setupConnector: func() *CommonConnector {
				c, err := New(
					context.Background(),
					"mongodb://localhost:27017",
					nil,
					WithUsername("admin"),
					WithPassword("secret"),
					WithAuthSource("admin"),
				)

				require.NoError(t, err)
				return c
			},
			filter:  bson.D{},
			wantErr: false,
			expected: &mongo.ListDatabasesResult{
				Databases: []mongo.DatabaseSpecification{
					{Name: "admin"},
					{Name: "config"},
					{Name: "local"},
				},
			},
		},
		{
			name: "invalid filter",
			setupConnector: func() *CommonConnector {
				c, err := New(
					context.Background(),
					"mongodb://localhost:27017",
					nil,
					WithUsername("admin"),
					WithPassword("secret"),
					WithAuthSource("admin"),
				)

				require.NoError(t, err)
				return c
			},
			wantErr: true,
		},
		{
			name: "nil client error",
			setupConnector: func() *CommonConnector {
				return &CommonConnector{
					client: nil,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := tt.setupConnector()
			databases, err := connector.ListDatabases(context.Background(), tt.filter)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				require.Len(t, databases.Databases, len(tt.expected.Databases))
			}
		})
	}
}

func TestListDatabaseNames(t *testing.T) {
	tests := []struct {
		name           string
		setupConnector func() *CommonConnector
		filter         bson.D
		wantErr        bool
		expected       []string
	}{
		{
			name: "success list database names",
			setupConnector: func() *CommonConnector {
				c, err := New(
					context.Background(),
					"mongodb://localhost:27017",
					nil,
					WithUsername("admin"),
					WithPassword("secret"),
					WithAuthSource("admin"),
				)

				require.NoError(t, err)
				return c
			},
			filter:  bson.D{},
			wantErr: false,
			expected: []string{
				"admin",
				"config",
				"local",
			},
		},
		{
			name: "nil client error",
			setupConnector: func() *CommonConnector {
				return &CommonConnector{
					client: nil,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := tt.setupConnector()
			names, err := connector.ListDatabaseNames(context.Background(), tt.filter)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				require.Equal(t, names, tt.expected)
			}
		})
	}
}
