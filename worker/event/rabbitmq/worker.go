package rabbitmq

// TODO: refactor to rabbitmq adapter

import (
	"fmt"
	"sync"

	"github.com/radianteam/framework/worker"

	"github.com/streadway/amqp"
)

type RabbitMqEventWorkerRegFunc func(d *amqp.Delivery, wc *worker.WorkerAdapters) error

type RabbitMqConfig struct {
	Host          string `json:"Host,omitempty" config:"Host,required"`
	Port          int16  `json:"Port,omitempty" config:"Port,required"`
	Username      string `json:"Username,omitempty" config:"Username,required"`
	Password      string `json:"Password,omitempty" config:"Password,required"`
	PrefetchCount int    `json:"PrefetchCount,omitempty" config:"PrefetchCount"`
}

type RabbitMqEventWorker struct {
	*worker.BaseWorker

	config *RabbitMqConfig

	connection *amqp.Connection

	mutex    sync.Mutex
	waitChan chan bool

	handlers map[string]map[string]RabbitMqEventWorkerRegFunc
}

func NewRabbitMqEventWorker(name string, config *RabbitMqConfig) *RabbitMqEventWorker {
	handlers := make(map[string]map[string]RabbitMqEventWorkerRegFunc)
	return &RabbitMqEventWorker{BaseWorker: worker.NewBaseWorker(name), config: config, handlers: handlers}
}

func (w *RabbitMqEventWorker) SetEvent(queue string, routingKey string, handler RabbitMqEventWorkerRegFunc) {
	if _, ok := w.handlers[queue]; !ok {
		w.handlers[queue] = make(map[string]RabbitMqEventWorkerRegFunc)
	}

	w.handlers[queue][routingKey] = handler
}

func (w *RabbitMqEventWorker) Setup() {
	w.Logger.Info("Setting up RabbitMq Events")
}

func (w *RabbitMqEventWorker) Run() {
	w.Logger.Info("Running RabbitMq Events")

	var err error = nil
	w.connection, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", w.config.Username, w.config.Password, w.config.Host, w.config.Port))

	if err != nil {
		w.Logger.Fatalf("dial %s\n", err)
	}

	wg := sync.WaitGroup{}

	for queueName, routingKeys := range w.handlers {
		wg.Add(1)

		go func(name string, handlers map[string]RabbitMqEventWorkerRegFunc) {
			defer wg.Done()

			w.Logger.Infof("Consuming queue %s", name)

			channel, err := w.connection.Channel()

			if err != nil {
				w.Logger.Fatalf("channel create %s\n", err)
			}

			err = channel.Qos(int(w.config.PrefetchCount), 0, false) // TODO: hardcode
			if err != nil {
				w.Logger.Fatalf("Cannot prepare qos - %s\n", err)
			}

			msgs, err := channel.Consume(name, w.GetName(), false, false, false, false, nil) // TODO: hardcoded values

			if err != nil {
				w.Logger.Fatalf("Consuming queue %s started with error %v", name, err)
			}

			for message := range msgs {
				w.Logger.Infof("Received a message from %s with key %s", name, message.RoutingKey)
				w.Logger.Debugf("Received message body: %s", message.Body)

				handler, ok := handlers[message.RoutingKey]

				if !ok {
					w.Logger.Errorf("Queue %s doesn't have a handler for %s routing key", name, message.RoutingKey)
					message.Acknowledger.Nack(message.DeliveryTag, false, false)

					continue
				}

				var err error

				//single thread processing. contexts can be none thread safe!
				w.mutex.Lock()
				err = handler(&message, w.Adapters)
				w.mutex.Unlock()

				if err != nil {
					w.Logger.Errorf("Queue %s routing key %s failed to proceed the message with delivery tag %d ", name, message.RoutingKey, message.DeliveryTag)
					message.Nack(true, false)

					continue
				}

				message.Ack(true)
			}

			channel.Cancel(name, false)
			channel.Close()

			w.Logger.Infof("Consuming queue %s stopped", name)
		}(queueName, routingKeys)
	}

	w.waitChan = make(chan bool)

	<-w.waitChan

	w.Logger.Info("Stopping RabbitMq Events")

	w.connection.Close()

	wg.Wait()
}

func (w *RabbitMqEventWorker) Stop() {
	w.Logger.Info("stop signal received! Graceful shutting down")

	close(w.waitChan)
}
