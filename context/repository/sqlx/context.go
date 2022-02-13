package sqlx

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type ContextSqlx struct {
	db *sqlx.DB

	driver            string
	connection_string string
}

func NewContextSqlx(driver string, connection_string string) *ContextSqlx {
	return &ContextSqlx{driver: driver, connection_string: connection_string}
}

func (c *ContextSqlx) Setup() error {
	var err error
	c.db, err = sqlx.Connect(c.driver, c.connection_string)

	return err
}

func (c *ContextSqlx) Close() error {
	return nil
}

func (c *ContextSqlx) Get() interface{} {
	return c.db
}

func (c *ContextSqlx) Begin() (*sqlx.Tx, error) {
	return c.db.Beginx()
}
