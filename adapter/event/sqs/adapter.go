package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/radianteam/framework/adapter"
)

type AwsSqsConfig struct {
	Endpoint            string `json:"Endpoint,omitempty,omitempty" config:"Endpoint"`
	Region              string `json:"Region,omitempty" config:"Region,required"`
	AccessKeyID         string `json:"AwsAccessKeyID,omitempty" config:"AwsAccessKeyID"`
	SecretAccessKey     string `json:"AwsSecretAccessKey,omitempty" config:"AwsSecretAccessKey"`
	SessionToken        string `json:"AwsSessionToken,omitempty" config:"AwsSessionToken"`
	MaxNumberOfMessages int64  `json:"MaxNumberOfMessages" config:"MaxNumberOfMessages"`
	WaitTimeSeconds     int64  `json:"WaitTimeSeconds" config:"WaitTimeSeconds"`
	VisibilityTimeout   int64  `json:"VisibilityTimeout" config:"VisibilityTimeout"`
	SharedCredentials   bool   `json:"SharedCredentials,omitempty" config:"SharedCredentials"`
}

type AwsSqsAdapter struct {
	*adapter.BaseAdapter

	config *AwsSqsConfig

	sess   *session.Session
	client *sqs.SQS
}

func NewAwsSqsAdapter(name string, config *AwsSqsConfig) *AwsSqsAdapter {
	return &AwsSqsAdapter{BaseAdapter: adapter.NewBaseAdapter(name), config: config}
}

func (a *AwsSqsAdapter) Setup() (err error) {
	endpoint := a.config.Endpoint
	cfg := aws.Config{
		Region:   aws.String(a.config.Region),
		Endpoint: aws.String(endpoint),
	}

	if !a.config.SharedCredentials {
		cfg.Credentials = credentials.NewStaticCredentials(a.config.AccessKeyID, a.config.SecretAccessKey, a.config.SessionToken)
	}

	a.sess = session.Must(session.NewSession(&cfg))
	a.client = sqs.New(a.sess)

	return
}

func (a *AwsSqsAdapter) Close() error {
	return nil
}

func (a *AwsSqsAdapter) CreateQueue(qName string) (err error) {
	createQueueInput := &sqs.CreateQueueInput{QueueName: aws.String(qName)}
	_, err = a.client.CreateQueue(createQueueInput)

	return
}

func (a *AwsSqsAdapter) ListQueues() (*[]*string, error) {
	result, err := a.client.ListQueues(&sqs.ListQueuesInput{})
	if err != nil {
		return nil, err
	}

	return &result.QueueUrls, nil
}

func (a *AwsSqsAdapter) getQueueUrl(qName string) (string, error) {
	getQueueUrlInput := &sqs.GetQueueUrlInput{QueueName: aws.String(qName)}
	result, err := a.client.GetQueueUrl(getQueueUrlInput)
	if err != nil {
		return "", err
	}

	return aws.StringValue(result.QueueUrl), nil
}

func (a *AwsSqsAdapter) DeleteQueue(qName string) (err error) {
	queueUrl, err := a.getQueueUrl(qName)
	if err != nil {
		return
	}

	deleteQueueInput := &sqs.DeleteQueueInput{QueueUrl: aws.String(queueUrl)}
	_, err = a.client.DeleteQueue(deleteQueueInput)

	return
}

func (a *AwsSqsAdapter) DeleteMessage(qName, receiptHandle string) (err error) {
	queueUrl, err := a.getQueueUrl(qName)
	if err != nil {
		return
	}

	deleteMessageInput := &sqs.DeleteMessageInput{QueueUrl: aws.String(queueUrl), ReceiptHandle: aws.String(receiptHandle)}
	_, err = a.client.DeleteMessage(deleteMessageInput)

	return
}

func (a *AwsSqsAdapter) Publish(qName, message string) (err error) {
	queueUrl, err := a.getQueueUrl(qName)
	if err != nil {
		return
	}

	sendMessageInput := &sqs.SendMessageInput{QueueUrl: aws.String(queueUrl), MessageBody: aws.String(message)}
	_, err = a.client.SendMessage(sendMessageInput)

	return
}

func (a *AwsSqsAdapter) Consume(queueUrl string) ([]*sqs.Message, error) {
	receiveMessageInput := &sqs.ReceiveMessageInput{QueueUrl: aws.String(queueUrl)}
	if a.config.MaxNumberOfMessages != 0 {
		receiveMessageInput.MaxNumberOfMessages = aws.Int64(a.config.MaxNumberOfMessages)
	}
	if a.config.WaitTimeSeconds != 0 {
		receiveMessageInput.WaitTimeSeconds = aws.Int64(a.config.WaitTimeSeconds)
	}
	if a.config.VisibilityTimeout != 0 {
		receiveMessageInput.VisibilityTimeout = aws.Int64(a.config.VisibilityTimeout)
	}
	res, err := a.client.ReceiveMessage(receiveMessageInput)
	if err != nil {
		return nil, err
	}

	return res.Messages, nil
}
