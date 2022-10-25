package s3

import (
	"io"
	"os"

	"github.com/radianteam/framework/adapter"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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

func (a *AwsS3Adapter) BucketItemUpload(bucket string, key string, body io.Reader) (err error) {
	uploader := s3manager.NewUploader(a.awsSession)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	})

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	return
}

func (a *AwsS3Adapter) BucketItemDownload(bucket string, key string, body io.WriterAt) (bytes int64, err error) {
	downloader := s3manager.NewDownloader(a.awsSession)

	bytes, err = downloader.Download(body,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	return
}

func (a *AwsS3Adapter) BucketItemDownloadBytes(bucket string, key string) ([]byte, error) {
	var b []byte
	buf := aws.NewWriteAtBuffer(b)

	_, err := a.BucketItemDownload(bucket, key, buf)

	// err logs is in BucketItemDownload

	return buf.Bytes(), err
}

func (a *AwsS3Adapter) BucketItemDownloadFile(bucket string, key string, path string) (numBytes int64, err error) {
	file, err := os.Create(path)

	if err != nil {
		logrus.WithField("adapter", a.GetName()).Error(err)
		return
	}

	defer file.Close()

	numBytes, err = a.BucketItemDownload(bucket, key, file)

	// err logs is in BucketItemDownload

	return numBytes, err
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

// TODO: implement restore item
// TODO: implement copy item
