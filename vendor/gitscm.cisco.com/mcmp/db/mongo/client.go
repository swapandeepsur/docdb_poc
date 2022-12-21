package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

	wraperrors "github.com/pkg/errors"
	"github.com/spf13/viper"
	"gitscm.cisco.com/mcmp/utils/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"gitscm.cisco.com/mcmp/db/mongo/config"
)

// Selector defines different modes the connection can be configured.
type Selector func(s *selections)

// AppName allows selecting an appName that is different from registered ServiceName.
func AppName(name string) Selector {
	return func(s *selections) {
		s.appName = name
	}
}

// ReadOnly selects a read-only connection.
func ReadOnly() Selector {
	return func(s *selections) {
		s.readOnly = true
	}
}

// ReadWrite selects a read-write connection.
func ReadWrite() Selector {
	return func(s *selections) {
		s.readWrite = true
	}
}

// UpgradeSchema selects a configuration appropriate for apply schema upgrades.
func UpgradeSchema() Selector {
	return func(s *selections) {
		s.upgrade = true
	}
}

type selections struct {
	readOnly  bool
	readWrite bool
	upgrade   bool
	appName   string
}

const (
	// Timeout operations after N seconds
	connectTimeout  = 5
	queryTimeout    = 30
	username        = ""
	password        = ""
	clusterEndpoint = "runondocumentdbpoc.cluster-cndxm9anwsxr.us-east-1.docdb.amazonaws.com:27017"

	// Which instances to read from
	readPreference           = "secondaryPreferred"
	connectionStringTemplate = "mongodb://%s:%s@%s/test?replicaSet=rs0&readpreference=%s"
)

// New uses the common configurations defined in config package to configuration a client
// for the specified database name.
func New(ctx context.Context, flag bool, picks ...Selector) (*mongo.Database, error) {
	if flag {
		connectionURI := fmt.Sprintf(connectionStringTemplate, username, password, clusterEndpoint, readPreference)

		client, err := mongo.NewClient(options.Client().ApplyURI(connectionURI))
		if err != nil {
			log(ctx).Infof("Failed to create client: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), connectTimeout*time.Second)
		defer cancel()

		err = client.Connect(ctx)
		if err != nil {
			log(ctx).Infof("Failed to connect to cluster: %v", err)
		}

		// Force a connection to verify our connection string
		err = client.Ping(ctx, nil)
		if err != nil {
			log(ctx).Infof("Failed to ping cluster: %v", err)
		}

		fmt.Println("Connected to DocumentDB!")

		return client.Database("test"), nil
	}else {
		if err := config.LoadMongoConfigs(); err != nil {
			return nil, wraperrors.Wrapf(err, "unable to load MongoDB configurations from %q", viper.GetString(config.MongoDBConfigFile))
		}

		s := applySelections(picks...)

		if viper.GetString(config.MongoDBClusterSrv) == "" && len(viper.GetStringSlice(config.MongoDBHosts)) == 0 {
			return nil, wraperrors.Errorf("missing both %q and %q configurations", config.MongoDBClusterSrv, config.MongoDBHosts)
		}

		if viper.GetString(config.MongoDBName) == "" {
			return nil, wraperrors.Errorf("missing %q configurations", config.MongoDBName)
		}

		tlsConfig, err := newTLSConfig()
		if err != nil {
			return nil, wraperrors.Wrap(err, "unable to create TLS config")
		}

		m := newMonitor(ctx, appName(s))

		opts := options.Client().
			SetAppName(appName(s)).
			SetReplicaSet(viper.GetString(config.MongoDBReplicaSet)).
			SetTLSConfig(tlsConfig).
			SetConnectTimeout(viper.GetDuration(config.MongoDBConnectTimeout)).
			SetSocketTimeout(socketTimeout(s)).
			SetServerSelectionTimeout(viper.GetDuration(config.MongoDBSelectTimeout)).
			SetMinPoolSize(uint64(1)).
			SetMaxPoolSize(uint64(viper.GetInt(config.MongoDBPoolLimit))).
			SetMaxConnIdleTime(viper.GetDuration(config.MongoDBPoolMaxIdleTime)).
			SetPoolMonitor(m.PoolMonitor()).
			SetReadPreference(readpref.PrimaryPreferred()).
			SetRetryReads(true).
			SetRetryWrites(true)

		if creds := newCredentials(); creds != nil {
			opts = opts.SetAuth(*creds)
		}

		c, err := mongo.Connect(ctx, connectTo(opts))
		if err != nil {
			return nil, wraperrors.Wrap(err, "unable to create mongo client")
		}

		if err := c.Ping(ctx, nil); err != nil {
			return nil, wraperrors.Wrap(err, "unable to reach mongo cluster")
		}

		return c.Database(viper.GetString(config.MongoDBName)), nil
	}
}

func applySelections(picks ...Selector) *selections {
	s := new(selections)

	for _, pick := range picks {
		pick(s)
	}

	if s.readOnly {
		config.ReadOnly()
	}

	if s.readWrite {
		config.ReadWrite()
	}

	return s
}

func appName(s *selections) string {
	if s.appName != "" {
		return s.appName
	}

	return viper.GetString(env.SvcName)
}

func socketTimeout(s *selections) time.Duration {
	if s.upgrade {
		return viper.GetDuration(config.MongoDBUpgradeTimeout)
	}

	return viper.GetDuration(config.MongoDBTimeout)
}

func connectTo(opts *options.ClientOptions) *options.ClientOptions {
	// default to DNS name for the cluster when available
	if viper.GetString(config.MongoDBClusterSrv) != "" {
		return opts.ApplyURI("mongodb+srv://" + viper.GetString(config.MongoDBClusterSrv))
	}

	return opts.SetHosts(viper.GetStringSlice(config.MongoDBHosts))
}

func newTLSConfig() (*tls.Config, error) {
	caFile := viper.GetString(config.MongoDBCACert)
	if caFile == "" {
		return nil, nil
	}

	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{RootCAs: caCertPool}, nil
}

func newCredentials() *options.Credential {
	// if the environment variables are defined with actual values
	// then configure the Credential with username and password
	if viper.GetString(config.MongoDBUsername) != "" && viper.GetString(config.MongoDBPassword) != "" {
		return &options.Credential{
			Username:   viper.GetString(config.MongoDBUsername),
			Password:   viper.GetString(config.MongoDBPassword),
			AuthSource: viper.GetString(config.MongoDBAuthSource),
		}
	}

	return nil
}
