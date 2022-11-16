package rabbitmq

import (
	"github.com/radianteam/framework/worker"
	"github.com/streadway/amqp"
)

type RabbitMqEventHandlerInterface interface {
	worker.BaseHandlerInterface

	SetMqMessage(*amqp.Delivery)
}

type RabbitMqEventHandler struct {
	worker.BaseHandler

	MqMessage *amqp.Delivery
}

func (h *RabbitMqEventHandler) SetMqMessage(m *amqp.Delivery) {
	h.MqMessage = m
}
