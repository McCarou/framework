package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/radianteam/framework/worker"
)

type AwsSqsEventHandlerInterface interface {
	worker.BaseHandlerInterface

	SetSqsMessage(*sqs.Message)
}

type AwsSqsEventHandler struct {
	worker.BaseHandler

	SqsMessage *sqs.Message
}

func (h *AwsSqsEventHandler) SetSqsMessage(m *sqs.Message) {
	h.SqsMessage = m
}
