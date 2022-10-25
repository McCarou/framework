package s3

import (
	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AwsS3Config struct {
	Endpoint        string `json:"endpoint,omitempty" config:"endpoint,required"`
	AccessKeyID     string `json:"access_key_id,omitempty" config:"access_key_id,required"`
	SecretAccessKey string `json:"secret_access_key,omitempty" config:"secret_access_key,required"`
	SessionToken    string `json:"session_token,omitempty" config:"session_token,required"`
	Region          string `json:"region,omitempty" config:"region,required"`
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
		Endpoint:    aws.String(a.config.Endpoint),
		Region:      aws.String(a.config.Region),
		Credentials: credentials.NewStaticCredentials(a.config.AccessKeyID, a.config.SecretAccessKey, a.config.SessionToken),
	})

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

func (a *AwsS3Adapter) BucketItemList(name string) ([]*s3.Object, error) {
	resp, err := a.s3Client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(name)})

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return nil, err

	}

	return resp.Contents, nil
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
