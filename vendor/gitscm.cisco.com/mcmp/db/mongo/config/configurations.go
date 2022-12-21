/*
Package config defines the supported configuration options for databases.

Example Configuration file (config.yaml)

	db.mongo:
		name: mytestdb
		username: mcmp_svc_rw
*/
package config

import (
	"os"

	wraperrors "github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// all configuration keys.
const (
	// Environment Variable: "MONGO_DB_HOSTS".
	MongoDBHosts = "db.mongo.hosts"
	// Environment Variable: "MONGO_DB_CLUSTER_SRV".
	MongoDBClusterSrv = "db.mongo.clustersrv"
	// Environment Variable: "MONGO_DB_NAME".
	MongoDBName = "db.mongo.name"
	// Environment Variable: "MONGO_DB_USERNAME".
	MongoDBUsername = "db.mongo.username"
	// Environment Variable: "MONGO_DB_PASSWORD".
	MongoDBPassword = "db.mongo.password.default"
	// Environment Variable: "MONGO_DB_AUTH_SOURCE".
	MongoDBAuthSource = "db.mongo.authsource"
	// Environment Variable: "MONGO_DB_REPLICASET".
	MongoDBReplicaSet = "db.mongo.replicaset"
	// Environment Variable: "MONGO_DB_CACERT".
	MongoDBCACert = "db.mongo.cacert"

	// Environment Variable: "MONGO_DB_CONFIG_FILE".
	MongoDBConfigFile = "db.mongo.configfile"

	// Default: "3s".
	MongoDBTimeout = "db.mongo.timeout.default"
	// Default: "3s".
	MongoDBConnectTimeout = "db.mongo.timeout.connect"
	// Default: "3s".
	MongoDBSelectTimeout = "db.mongo.timeout.select"
	// Default: "5m".
	MongoDBUpgradeTimeout = "db.mongo.timeout.upgrade"
	// Default: "2s".
	MongoDBIndexTimeout = "db.mongo.timeout.index"

	// Default: 10.
	MongoDBPoolLimit = "db.mongo.pool.limit"
	// Default: "5m".
	MongoDBPoolMaxIdleTime = "db.mongo.pool.maxidle"
)

const (
	// default path to current configurations.
	currentCfgFile = "/opt/mcmp/db/mongo/current.yaml"
	// default path to future configurations.
	futureCfgFile = "/opt/mcmp/db/mongo/future.yaml"

	mongoDBPasswordRW = "db.mongo.password.read_write"
	mongoDBPasswordRO = "db.mongo.password.read_only"
)

func init() {
	initialize()
}

// Current sets configuration file to use the current configurations.
// This is useful when needing to switch from testing a future configurations back to current.
func Current() {
	viper.Set(MongoDBConfigFile, currentCfgFile)
}

// Future sets configuration file to use the future configurations.
// This is useful when needing to switch to using a future configuration file.
func Future() {
	viper.Set(MongoDBConfigFile, futureCfgFile)
}

// ReadOnly selects the read only password.
func ReadOnly() {
	viper.Set(MongoDBPassword, viper.GetString(mongoDBPasswordRO))
}

// ReadWrite selects the read write password. Default configuration.
func ReadWrite() {
	viper.Set(MongoDBPassword, viper.GetString(mongoDBPasswordRW))
}

// LoadMongoConfigs loads a specified MongoDB configuration file and merges the content into existing sets of configurations.
func LoadMongoConfigs() error {
	name := viper.GetString(MongoDBConfigFile)
	if name == "" {
		return wraperrors.New("no MongoDB configuration file found")
	}

	f, err := os.Open(name)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	var cfg map[string]interface{}
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return err
	}

	return viper.MergeConfigMap(cfg)
}

// Restore will reset viper and re-initialize back to the default configurations
// For Testing ONLY!
func Restore() {
	viper.Reset()
	initialize()
}

// LoadFixture will load test fixture configuration; for testing only!
func LoadFixture(dir string) error {
	viper.SetConfigName("config")
	viper.AddConfigPath(dir)

	return viper.ReadInConfig()
}

func initialize() {
	viper.SetDefault(MongoDBConfigFile, currentCfgFile)
	viper.SetDefault(MongoDBTimeout, "30s")
	viper.SetDefault(MongoDBConnectTimeout, "30s")
	viper.SetDefault(MongoDBSelectTimeout, "30s")
	viper.SetDefault(MongoDBUpgradeTimeout, "5m")
	viper.SetDefault(MongoDBIndexTimeout, "2s")
	viper.SetDefault(MongoDBPoolLimit, 10)
	viper.SetDefault(MongoDBPoolMaxIdleTime, "15m")

	_ = viper.BindEnv(MongoDBHosts, "MONGO_DB_HOSTS")
	_ = viper.BindEnv(MongoDBClusterSrv, "MONGO_DB_CLUSTER_SRV")
	_ = viper.BindEnv(MongoDBName, "MONGO_DB_NAME")
	_ = viper.BindEnv(MongoDBUsername, "MONGO_DB_USERNAME")
	_ = viper.BindEnv(MongoDBPassword, "MONGO_DB_PASSWORD")
	_ = viper.BindEnv(MongoDBAuthSource, "MONGO_DB_AUTH_SOURCE")
	_ = viper.BindEnv(MongoDBReplicaSet, "MONGO_DB_REPLICASET")
	_ = viper.BindEnv(MongoDBCACert, "MONGO_DB_CACERT")
	_ = viper.BindEnv(MongoDBConfigFile, "MONGO_DB_CONFIG_FILE")
}
