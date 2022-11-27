package mongodb

// TODO: load cert either from a file or from a config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"

	"github.com/radianteam/framework/adapter"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDbConfig struct {
	Hosts            []string `json:"Hosts,omitempty" config:"Hosts,required"`
	Username         string   `json:"Username,omitempty" config:"Username"`
	Password         string   `json:"Password,omitempty" config:"Password"`
	ReplicaSet       string   `json:"ReplicaSet,omitempty" config:"ReplicaSet"`
	DirectConnection bool     `json:"DirectConnection,omitempty" config:"DirectConnection"`
	RootCA           string   `json:"RootCA,omitempty" config:"RootCA"`
	AuthSource       string   `json:"AuthSource,omitempty" config:"AuthSource"`
}

type MongoDbAdapter struct {
	*adapter.BaseAdapter

	client *mongo.Client

	config *MongoDbConfig
}

func NewMongoDbAdapter(name string, config *MongoDbConfig) *MongoDbAdapter {
	return &MongoDbAdapter{BaseAdapter: adapter.NewBaseAdapter(name), config: config}
}

func (a *MongoDbAdapter) Setup() (err error) {
	mongoOpt := options.Client()

	mongoOpt.SetHosts(a.config.Hosts)

	if strings.TrimSpace(a.config.Username) != "" || strings.TrimSpace(a.config.Password) != "" {
		mongoOpt.SetAuth(options.Credential{Username: a.config.Username, Password: a.config.Password, AuthSource: a.config.AuthSource})
	}

	if strings.TrimSpace(a.config.ReplicaSet) != "" {
		mongoOpt.SetReplicaSet(a.config.ReplicaSet)
	}

	mongoOpt.SetDirect(a.config.DirectConnection)

	if strings.TrimSpace(a.config.RootCA) != "" {
		rootCerts := x509.NewCertPool()

		if ca, err := os.ReadFile(a.config.RootCA); err == nil {
			rootCerts.AppendCertsFromPEM(ca)
		} else {
			return err
		}

		mongoOpt.SetTLSConfig(&tls.Config{RootCAs: rootCerts})
	}

	a.client, err = mongo.Connect(context.TODO(), mongoOpt)

	if err != nil {
		a.Logger.Error(err)
	}

	return
}

func (a *MongoDbAdapter) Close() error {
	return a.client.Disconnect(context.TODO())
}

func (a *MongoDbAdapter) Get() interface{} {
	return a.client
}

func (a *MongoDbAdapter) GetDatabase(name string) *mongo.Database {
	return a.client.Database(name)
}

func (a *MongoDbAdapter) GetCollection(dbName string, collName string) *mongo.Collection {
	return a.client.Database(dbName).Collection(collName)
}

func (a *MongoDbAdapter) CreateSession() (mongo.Session, error) {
	return a.client.StartSession()
}
