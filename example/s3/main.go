package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/radianteam/framework/adapter/storage/s3"
)

func main() {
	// setup credentials for the S3 adapter
	s3Adapter := s3.NewAwsS3Adapter("s3", &s3.AwsS3Config{Region: "us-west-2", Endpoint: "http://s3.localhost.localstack.cloud:4566", AccessKeyID: "test", SecretAccessKey: "hat", SessionToken: "hat2"})
	//s3Adapter := s3.NewAwsS3Adapter("s3", &s3.AwsS3Config{Region: "us-west-2", SharedCredentialFile: true})

	s3Adapter.Setup()

	// create S3 bucket
	err := s3Adapter.BucketCreate("testbucket", true)

	if err != nil {
		fmt.Printf("error: %v", err)
	} else {
		fmt.Println("bucket created")
	}

	// get the list of buckets
	blist, err := s3Adapter.BucketList()

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	for _, s := range blist {
		fmt.Printf("bucket: %s\n", s)
	}

	// put a new file into the backet
	rdr := strings.NewReader("Hello, world! :)")
	err = s3Adapter.BucketItemUpload("testbucket", "testfolder/text.txt", rdr)

	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Println("item uploaded")
	}

	// get the list of items
	bilist, err := s3Adapter.BucketItemList("testbucket")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	for _, s := range bilist {
		fmt.Printf("bucket item: %s %s %d %s\n", *s.Key, s.LastModified.Format(time.RFC3339Nano), s.Size, *s.StorageClass)
	}

	// get a file content from the bucket
	bytes, err := s3Adapter.BucketItemDownloadBytes("testbucket", "testfolder/text.txt")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Printf("bucket item content: %s\n", string(bytes))

	// download a file from the bucket
	bread, err := s3Adapter.BucketItemDownloadFile("testbucket", "testfolder/text.txt", "dtext.txt")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Printf("bytes read: %d\n", bread)

	// clear the bucket
	err = s3Adapter.BucketClear("testbucket")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Println("bucket cleared")
	}

	// delete the bucket
	err = s3Adapter.BucketDelete("testbucket", true)

	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Println("bucket deleted")
	}

	// close the adapter
	s3Adapter.Close()
}
