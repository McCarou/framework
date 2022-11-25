package rabbitmq

import (
	"errors"
	"fmt"
	"time"

	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	RabbitMqPublishTimeoutMs = 5000 // TODO: make configurable
	NotifyChannelSize        = 10
)

type RabbitMqConfig struct {
	Host     string `json:"Host,omitempty" config:"Host,required"`
	Port     uint16 `json:"Port,omitempty" config:"Port,required"`
	Username string `json:"Username,omitempty" config:"Username"`
	Password string `json:"Password,omitempty" config:"Password"`
	Exchange string `json:"Exchange,omitempty" config:"Exchange"`
}

type RabbitMqAdapter struct {
	*adapter.BaseAdapter

	config *RabbitMqConfig

	connection        *amqp.Connection
	channel           *amqp.Channel
	notifyPublishChan chan amqp.Confirmation
	notifyCloseChan   chan *amqp.Error
}

func NewRabbitMqAdapter(name string, config *RabbitMqConfig) *RabbitMqAdapter {
	return &RabbitMqAdapter{BaseAdapter: adapter.NewBaseAdapter(name), config: config}
}

func (a *RabbitMqAdapter) setupChannel() (err error) {
	a.channel, err = a.connection.Channel()
	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	a.channel.Confirm(false)
	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	a.notifyPublishChan = make(chan amqp.Confirmation, NotifyChannelSize)
	a.notifyPublishChan = a.channel.NotifyPublish(a.notifyPublishChan)

	a.notifyCloseChan = make(chan *amqp.Error, NotifyChannelSize)
	a.notifyCloseChan = a.channel.NotifyClose(a.notifyCloseChan)

	return a.channel.Qos(1, 0, false) // TODO: hardcode
}

func (a *RabbitMqAdapter) Setup() (err error) {
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

	return a.setupChannel()
}

func (a *RabbitMqAdapter) Close() (err error) {
	if err = a.channel.Close(); err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	return a.connection.Close()
}

func (a *RabbitMqAdapter) checkConnection() (err error) {
	if a.connection.IsClosed() {
		return a.Setup()
	}

	select {
	case x, ok := <-a.notifyCloseChan:
		if ok {
			logrus.WithField("adapter", a.GetName()).Debug("AMQP channel closed with error: %v", x)
		} else {
			logrus.WithField("adapter", a.GetName()).Debug("AMQP channel closed!")
		}

		return a.setupChannel()
	default:
	}

	return

}

func (a *RabbitMqAdapter) DeclareExchange(name string, kind string, durable bool) (err error) {
	if err = a.checkConnection(); err != nil {
		return
	}

	return a.channel.ExchangeDeclare(name, kind, durable, false, false, false, nil)
}

func (a *RabbitMqAdapter) DeclareQueue(name string, durable bool) (err error) {
	if err = a.checkConnection(); err != nil {
		return
	}

	_, err = a.channel.QueueDeclare(name, durable, false, false, false, nil)

	return err
}

func (a *RabbitMqAdapter) BindQueue(exchange string, routingKey string, queue string) (err error) {
	if err = a.checkConnection(); err != nil {
		return
	}

	err = a.channel.QueueBind(queue, routingKey, exchange, false, nil)

	return err
}

func (a *RabbitMqAdapter) PublishExchange(exchange string, key string, message []byte) (err error) {
	if err = a.checkConnection(); err != nil {
		return
	}

	err = a.channel.Publish(exchange, key, false, false, amqp.Publishing{Body: message}) // TODO: connection can be ok but channel is closed

	if err != nil {
		return err
	}

	var confirmation amqp.Confirmation

	select {
	case confirmation = <-a.notifyPublishChan:
	case <-time.After(time.Millisecond * RabbitMqPublishTimeoutMs):
		return errors.New("publishing error: timeout")
	}

	if confirmation.Ack {
		return
	} else {
		return errors.New("publishing error: wrong confirmation")
	}
}

func (a *RabbitMqAdapter) Publish(key string, message []byte) (err error) {
	return a.PublishExchange(a.config.Exchange, key, message)
}
