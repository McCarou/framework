package rabbitmq

import (
	"fmt"
	"sync"

	"github.com/radianteam/framework/worker"

	"github.com/streadway/amqp"
)

type EventRabbitMq struct {
	*worker.WorkerBase

	connection *amqp.Connection
	channel    *amqp.Channel

	mutex     sync.Mutex
	wait_chan chan bool

	handlers map[string]map[string]func(d *amqp.Delivery, wc *worker.WorkerContexts) error
}

func NewEventRabbitMq(config *worker.WorkerConfig) *EventRabbitMq {
	handlers := make(map[string]map[string]func(d *amqp.Delivery, wc *worker.WorkerContexts) error)
	return &EventRabbitMq{WorkerBase: worker.NewWorkerBase(config), handlers: handlers}
}

func (w *EventRabbitMq) SetEvent(queue string, routing_key string, handler func(d *amqp.Delivery, wc *worker.WorkerContexts) error) {
	_, ok := w.handlers[queue]

	if !ok {
		w.handlers[queue] = (make(map[string]func(d *amqp.Delivery, wc *worker.WorkerContexts) error))
	}

	w.handlers[queue][routing_key] = handler
}

func (w *EventRabbitMq) Setup() {
	w.Logger.Info("Setting up RabbitMq Events")
}

func (w *EventRabbitMq) Run() {
	w.Logger.Info("Running RabbitMq Events")

	var err error = nil
	w.connection, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", w.Config.Login, w.Config.Password, w.Config.Host, w.Config.Port))

	if err != nil {
		w.Logger.Fatalf("dial %s\n", err)
	}

	w.channel, err = w.connection.Channel()

	if err != nil {
		w.Logger.Fatalf("channel create %s\n", err)
	}

	w.channel.Qos(1, 0, false)

	wg := sync.WaitGroup{}

	for queue_name, routing_keys := range w.handlers {
		wg.Add(1)

		go func(name string, handlers map[string]func(d *amqp.Delivery, wc *worker.WorkerContexts) error) {
			defer wg.Done()

			w.Logger.Infof("Consuming queue %s", name)

			msgs, err := w.channel.Consume(
				name,  // queue
				name,  // consumer
				false, // auto-ack
				false, // exclusive
				false, // no-local
				false, // no-wait
				nil,   // args
			)

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
				err = handler(&message, w.Contexts)
				w.mutex.Unlock()

				if err != nil {
					w.Logger.Errorf("Queue %s routing key %s failed to proceed the message with delivery tag %d ", name, message.RoutingKey, message.DeliveryTag)
					message.Nack(true, false)

					continue
				}

				message.Ack(true)
			}

			w.Logger.Infof("Consuming queue %s stopped", name)
		}(queue_name, routing_keys)
	}

	w.wait_chan = make(chan bool)

	<-w.wait_chan

	w.Logger.Info("Stopping RabbitMq Events")

	for queue_name := range w.handlers {
		w.channel.Cancel(queue_name, false)
	}

	w.channel.Close()
	w.connection.Close()

	wg.Wait()
}

func (w *EventRabbitMq) Stop() {
	w.Logger.Info("stop signal received! Graceful shutting down")

	close(w.wait_chan)
}
