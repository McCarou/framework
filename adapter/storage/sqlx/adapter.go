package adapter

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"
)

type SqlxConfig struct {
	Driver           string `json:"driver" config:"driver,required"`
	ConnectionString string `json:"connection_string,omitempty" config:"connection_string,required"`
}

type SqlxAdapter struct {
	*adapter.BaseAdapter

	config *SqlxConfig

	db *sqlx.DB
}

func NewSqlxAdapter(name string, config *SqlxConfig) *SqlxAdapter {
	return &SqlxAdapter{BaseAdapter: adapter.NewBaseAdapter(name), config: config}
}

func (a *SqlxAdapter) Setup() (err error) {
	a.db, err = sqlx.Connect(a.config.Driver, a.config.ConnectionString)
	if err == nil {
		// TODO: hardcode!
		a.db.SetMaxIdleConns(0)
		a.db.SetMaxOpenConns(1)
	} else {
		logrus.WithField("adapter", a.GetName()).Error(err)
	}

	return
}

func (a *SqlxAdapter) Close() (err error) {
	return
}

func (a *SqlxAdapter) Get() *sqlx.DB {
	return a.db
}

func (c *SqlxAdapter) Begin() (*sqlx.Tx, error) {
	return c.db.Beginx()
}
