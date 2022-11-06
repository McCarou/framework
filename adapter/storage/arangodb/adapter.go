package arangodb

import (
	"context"
	"crypto/tls"

	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

type ArangoDbConfig struct {
	Servers            []string `json:"Servers,omitempty" config:"Servers,required"`
	Username           string   `json:"Username,omitempty" config:"Username,required"`
	Password           string   `json:"Password,omitempty" config:"Password"`
	Database           string   `json:"Database,omitempty" config:"Database,required"`
	InsecureSkipVerify bool     `json:"InsecureSkipVerify,omitempty" config:"InsecureSkipVerify"`
}

type ArangoDbAdapter struct {
	*adapter.BaseAdapter

	config *ArangoDbConfig

	database driver.Database
}

func NewArangoDbAdapter(name string, config *ArangoDbConfig) *ArangoDbAdapter {
	return &ArangoDbAdapter{BaseAdapter: adapter.NewBaseAdapter(name), config: config}
}

func (a *ArangoDbAdapter) Setup() (err error) {
	connConfig := http.ConnectionConfig{Endpoints: a.config.Servers}

	if a.config.InsecureSkipVerify {
		connConfig.TLSConfig = &tls.Config{InsecureSkipVerify: a.config.InsecureSkipVerify}
	}

	conn, err := http.NewConnection(connConfig)
	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	client, err := driver.NewClient(driver.ClientConfig{Connection: conn, Authentication: driver.BasicAuthentication(a.config.Username, a.config.Password)})
	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	if ok, _ := client.DatabaseExists(context.TODO(), a.config.Database); !ok {
		a.database, err = client.CreateDatabase(context.TODO(), a.config.Database, nil)
		if err != nil {
			logrus.WithField("adapter", a.GetName()).Error(err)
			return
		}
	}

	a.database, err = client.Database(context.TODO(), a.config.Database)
	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
	}
	return
}

func (a *ArangoDbAdapter) Close() error {
	return nil
}

func (a *ArangoDbAdapter) Get() driver.Database {
	return a.database
}
