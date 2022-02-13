package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ContextMongoDb struct {
	client *mongo.Client

	connection_string string
}

func NewContextMongoDb(connection_string string) *ContextMongoDb {
	return &ContextMongoDb{connection_string: connection_string}
}

func (c *ContextMongoDb) Setup() error {
	var err error
	c.client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(c.connection_string))

	return err
}

func (c *ContextMongoDb) Close() error {
	return c.client.Disconnect(context.TODO())
}

func (c *ContextMongoDb) Get() interface{} {
	return c.client
}

func (c *ContextMongoDb) GetDatabase(name string) *mongo.Database {
	return c.client.Database(name)
}

func (c *ContextMongoDb) GetCollection(db_name string, coll_name string) *mongo.Collection {
	return c.client.Database(db_name).Collection(coll_name)
}
