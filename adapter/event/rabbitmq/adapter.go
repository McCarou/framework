package rmqa

import (
	"fmt"
	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type RabbitMqConfig struct {
	Host     string   `json:"host,omitempty" config:"host,required"`
	Port     int16    `json:"port,omitempty" config:"port,required"`
	Username string   `json:"username,omitempty" config:"username"`
	Password string   `json:"password,omitempty" config:"password"`
	Exchange string   `json:"exchange,omitempty" config:"exchange"`
	Listen   []string `json:"listen,omitempty" config:"listen"`
}

type AdapterRabbitMq struct {
	*adapter.BaseAdapter

	config *RabbitMqConfig

	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewAdapterRabbitMq(name string, config *RabbitMqConfig) *AdapterRabbitMq {
	return &AdapterRabbitMq{BaseAdapter: adapter.NewBaseAdapter(name), config: config}
}

func (a *AdapterRabbitMq) Setup() (err error) {
	connStr := ""
	if a.config.Username != "" {
		connStr += a.config.Username

		if a.config.Password != "" {
			connStr += ":" + a.config.Password
		}

		connStr += "@"
	}
	connStr = fmt.Sprintf("amqp://%s%s:%d/", connStr, a.config.Host, a.config.Port)

	a.connection, err = amqp.Dial(connStr)
	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	a.channel, err = a.connection.Channel()
	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	return a.channel.Qos(1, 0, false)
}
func (a *AdapterRabbitMq) Close() (err error) {
	if err = a.channel.Close(); err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	return a.connection.Close()
}

func (a *AdapterRabbitMq) checkConnection() (err error) {
	if !a.connection.IsClosed() {
		return
	}

	return a.Setup()
}

func (a *AdapterRabbitMq) DeclareExchange(name string, kind string, durable bool) (err error) {
	if err = a.checkConnection(); err != nil {
		return
	}

	return a.channel.ExchangeDeclare(name, kind, durable, false, false, false, nil)
}

func (a *AdapterRabbitMq) PublishExchange(exchange string, key string, message []byte) (err error) {
	if err = a.checkConnection(); err != nil {
		return
	}

	return a.channel.Publish(exchange, key, false, false, amqp.Publishing{Body: message})
}

func (a *AdapterRabbitMq) Publish(key string, message []byte) (err error) {
	if err = a.checkConnection(); err != nil {
		return
	}

	return a.channel.Publish(a.config.Exchange, key, false, false, amqp.Publishing{Body: message})
}
