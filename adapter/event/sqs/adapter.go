package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/radianteam/framework/adapter"
)

type SqsConfig struct {
	Host            string `json:"host,omitempty" config:"host,required"`
	Port            int16  `json:"port,omitempty" config:"port,required"`
	Region          string `json:"region,omitempty" config:"region,required"`
	AccessKeyId     string `json:"awsAccessKeyId,omitempty" config:"awsAccessKeyId,required"`
	SecretAccessKey string `json:"awsSecretAccessKey,omitempty" config:"awsSecretAccessKey,required"`
	SessionToken    string `json:"awsSessionToken,omitempty" config:"awsSessionToken,required"`
	PollingTime     int    `json:"pollingTime,omitempty" config:"pollingTime,required"`
}

type SqsAdapter struct {
	*adapter.BaseAdapter

	config *SqsConfig

	session *session.Session
	client  *sqs.SQS
}

func NewSqsAdapter(name string, config *SqsConfig) *SqsAdapter {
	return &SqsAdapter{BaseAdapter: adapter.NewBaseAdapter(name), config: config}
}

func (a *SqsAdapter) Setup() (err error) {
	endpoint := fmt.Sprintf("http://%s:%d/", a.config.Host, a.config.Port)
	cfg := aws.Config{
		Region:      aws.String(endpoints.UsEast1RegionID),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(a.config.AccessKeyId, a.config.SecretAccessKey, a.config.SessionToken),
	}

	a.session = session.Must(session.NewSession(&cfg))
	a.client = sqs.New(a.session)

	return
}

func (a *SqsAdapter) Close() error {
	return nil
}

func (a *SqsAdapter) CreateQueue(input *sqs.CreateQueueInput) (err error) {
	_, err = a.client.CreateQueue(input)
	if err != nil {
		return
	}

	return
}

func (a *SqsAdapter) ListQueues(input *sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
	result, err := a.client.ListQueues(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (a *SqsAdapter) GetQueueUrl(input *sqs.GetQueueUrlInput) (string, error) {
	result, err := a.client.GetQueueUrl(input)
	if err != nil {
		return "", err
	}

	return aws.StringValue(result.QueueUrl), nil
}

func (a *SqsAdapter) DeleteQueue(input *sqs.DeleteQueueInput) (err error) {
	_, err = a.client.DeleteQueue(input)

	return
}

func (a *SqsAdapter) DeleteMessage(input *sqs.DeleteMessageInput) (err error) {
	_, err = a.client.DeleteMessage(input)
	if err != nil {
		return
	}

	return
}

func (a *SqsAdapter) Publish(input *sqs.SendMessageInput) (err error) {
	_, err = a.client.SendMessage(input)

	return
}

func (a *SqsAdapter) Consume(input *sqs.ReceiveMessageInput) ([]*sqs.Message, error) {
	res, err := a.client.ReceiveMessage(input)

	if err != nil {
		return nil, err
	}

	return res.Messages, nil
}
