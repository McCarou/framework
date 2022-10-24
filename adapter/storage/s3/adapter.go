package s3

import (
	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AwsS3Config struct {
	Servers            []string `json:"servers,omitempty" config:"servers,required"`
	Username           string   `json:"username,omitempty" config:"username,required"`
	Password           string   `json:"password,omitempty" config:"password"`
	Database           string   `json:"database,omitempty" config:"database,required"`
	InsecureSkipVerify bool     `json:"insecure_skip_verify,omitempty" config:"insecure_skip_verify"`
}

type AwsS3Adapter struct {
	*adapter.BaseAdapter

	config *AwsS3Config

	awsSession *session.Session
	s3Client   *s3.S3
}

func NewAwsS3Adapter(name string, config *AwsS3Config) *AwsS3Adapter {
	return &AwsS3Adapter{BaseAdapter: adapter.NewBaseAdapter(name), config: config}
}

func (a *AwsS3Adapter) Setup() (err error) {
	a.awsSession, err = session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	// Create S3 service client
	a.s3Client = s3.New(a.awsSession)

	return
}

func (a *AwsS3Adapter) Close() error {
	return nil
}

func (a *AwsS3Adapter) BucketList() (response []string, err error) {
	result, err := a.s3Client.ListBuckets(nil)

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	for _, b := range result.Buckets {
		response = append(response, aws.StringValue(b.Name))
	}

	return
}

func (a *AwsS3Adapter) BucketCreate(name string, wait bool) (err error) {
	_, err = a.s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(name),
	})

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	if wait {
		err = a.s3Client.WaitUntilBucketExists(&s3.HeadBucketInput{
			Bucket: aws.String(name),
		})

		if err != nil {
			logrus.WithField("adapter", a.GetName()).Error(err)
			return
		}
	}

	return
}

func (a *AwsS3Adapter) BucketItemList() *s3.S3 {
	return a.s3Client
}

func (a *AwsS3Adapter) BucketItemUpload() *s3.S3 {
	return a.s3Client
}

func (a *AwsS3Adapter) BucketItemDownload() *s3.S3 {
	return a.s3Client
}

func (a *AwsS3Adapter) BucketItemDelete() *s3.S3 {
	return a.s3Client
}

func (a *AwsS3Adapter) BucketClear() *s3.S3 {
	return a.s3Client
}

func (a *AwsS3Adapter) BucketDelete() *s3.S3 {
	return a.s3Client
}
