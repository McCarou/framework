package rabbitmq

import (
	"fmt"

	"github.com/streadway/amqp"
)

type RabbitMqConfig struct {
	Host     string
	Port     int16
	Login    string
	Password string
	Exchange string
}

type ContextRabbitMq struct {
	config *RabbitMqConfig

	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewContextRabbitMq(config *RabbitMqConfig) *ContextRabbitMq {
	return &ContextRabbitMq{config: config}
}

func (c *ContextRabbitMq) Setup() error {
	var err error = nil

	c.connection, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", c.config.Login, c.config.Password, c.config.Host, c.config.Port))

	if err != nil {
		return err
	}

	c.channel, err = c.connection.Channel()

	if err != nil {
		return err
	}

	c.channel.Qos(1, 0, false)

	return nil
}

func (c *ContextRabbitMq) Close() error {
	c.channel.Close()
	c.connection.Close()

	return nil
}

func (c *ContextRabbitMq) Get() interface{} {
	return nil
}

func (c *ContextRabbitMq) DeclareExchange(name string, type_ string, durable bool) error {

	if c.connection.IsClosed() {
		err := c.Setup()

		if err != nil {
			return err
		}
	}

	err := c.channel.ExchangeDeclare(
		name,    // name
		type_,   // type
		durable, // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)

	return err
}

func (c *ContextRabbitMq) PublishExchange(exchange string, routing_key string, message []byte) error {

	if c.connection.IsClosed() {
		err := c.Setup()

		if err != nil {
			return err
		}
	}

	return c.channel.Publish(exchange, routing_key, false, false, amqp.Publishing{Body: message})
}

func (c *ContextRabbitMq) Publish(routing_key string, message []byte) error {

	if c.connection.IsClosed() {
		err := c.Setup()

		if err != nil {
			return err
		}
	}

	return c.channel.Publish(c.config.Exchange, routing_key, false, false, amqp.Publishing{Body: message})
}
