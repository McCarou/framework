# Example: AWS S3 adapter

1. [`Manual`](#1-manual)
2. [`Docker compose`](#2-docker-compose)

## 1 Manual

Create new folder. Create go.mod file inside with the following content:

``` go
module example

go 1.19

require (
	github.com/radianteam/framework v0.3.0
)
```

This file declares a module and the framework requirement.

Create a file named main.go and define the package inside:

``` go
package main
```

Create a main function and create an instance of the S3 adapter with static credentials (this example uses [localstack](https://github.com/localstack)):

``` go
func main() {
	// setup credentials for the S3 adapter
	s3Adapter := s3.NewAwsS3Adapter("s3", &s3.AwsS3Config{Region: "us-west-2", Endpoint: "http://s3.localhost.localstack.cloud:4566", AccessKeyID: "test", SecretAccessKey: "hat", SessionToken: "hat2"})
```

or decrale using a shared credential if it presents in environment variables or in ~.aws/ directory:

``` go
	s3Adapter := s3.NewAwsS3Adapter("s3", &s3.AwsS3Config{Region: "us-west-2", SharedCredentials: true})
```

Setup the adapter:

``` go
	s3Adapter.Setup()
```

Bucket creation example:

``` go
	// create S3 bucket
	err := s3Adapter.BucketCreate("testbucket", true)

	if err != nil {
		fmt.Printf("error: %v", err)
	} else {
		fmt.Println("bucket created")
	}
```

Bucket list example:

``` go
	// get the list of buckets
	blist, err := s3Adapter.BucketList()

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	for _, s := range blist {
		fmt.Printf("bucket: %s\n", s)
	}
```

Bucket file uploading example:

``` go
	// put a new file into the backet
	rdr := strings.NewReader("Hello, world! :)")
	err = s3Adapter.BucketItemUpload("testbucket", "testfolder/text.txt", rdr)

	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Println("item uploaded")
	}
```

Bucket item list example:

``` go
	// get the list of items
	bilist, err := s3Adapter.BucketItemList("testbucket", "")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	for _, s := range bilist {
		fmt.Printf("bucket item: %s %s %d %s\n", *s.Key, s.LastModified.Format(time.RFC3339Nano), s.Size, *s.StorageClass)
	}
```

Bucket item list with prefix example:

``` go
	// get the list of items in a folder
	bilist, err = s3Adapter.BucketItemList("testbucket", "testfolder")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	for _, s := range bilist {
		fmt.Printf("bucket item: %s %s %d %s\n", *s.Key, s.LastModified.Format(time.RFC3339Nano), s.Size, *s.StorageClass)
	}
```

Bucket item download to string example:

``` go
	// get a file content from the bucket
	bytes, err := s3Adapter.BucketItemDownloadBytes("testbucket", "testfolder/text.txt")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Printf("bucket item content: %s\n", string(bytes))
```

Bucket item download to a file example:

``` go
	// download a file from the bucket
	bread, err := s3Adapter.BucketItemDownloadFile("testbucket", "testfolder/text.txt", "dtext.txt")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Printf("bytes read: %d\n", bread)
```

Bucket clearing example:

``` go
	// clear the bucket
	err = s3Adapter.BucketClear("testbucket")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Println("bucket cleared")
	}
```

Bucket remove example:

``` go
	// delete the bucket
	err = s3Adapter.BucketDelete("testbucket", true)

	if err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		fmt.Println("bucket deleted")
	}
```

Close the adapter:

``` go
	// close the adapter
	s3Adapter.Close()
}
```
<br>

Now run the following command to download all requriments and prepare the application to start:

```
go mod tidy
```

Wait for the requirements download and then run the app:

```
go run main.go
```

Example
```
radian@radian s3 % go run main.go                                 
bucket created
bucket: testbucket
item uploaded
bucket item: testfolder/text.txt 2022-11-04T19:49:39Z 274878637016 STANDARD
bucket item: testfolder/text.txt 2022-11-04T19:49:39Z 274879109464 STANDARD
bucket item content: Hello, world! :)
bytes read: 16
bucket cleared
bucket deleted
```

If something goes wrong check [`main.go`](main.go) file or play with it in containers.

<br>

## 2 Docker compose

WARNING: you must have docker and docker-compose installed on your system. Use [`this instruction`](https://docs.docker.com/compose/install/) if you don't have it.

### 1 Clone the repository

```
git clone https://github.com/radianteam/framework.git
```
```
cd framework
```

### 2 Goto this folder

```
cd example/s3
```


### 3 Run the application

```
docker-compose up -d | grep app
```

### 4 Watch the results
```
Creating localstack_main ... done
Creating s3_app_1        ... done
Attaching to localstack_main, s3_app_1
app_1         | bucket created
app_1         | bucket: testbucket
app_1         | item uploaded
app_1         | bucket item: testfolder/text.txt 2022-11-04T20:13:34Z 274882193272 STANDARD
app_1         | bucket item: testfolder/text.txt 2022-11-04T20:13:34Z 274882195016 STANDARD
app_1         | bucket item content: Hello, world! :)
app_1         | bytes read: 16
app_1         | bucket cleared
app_1         | bucket deleted
```

### 5 Enjoy!

And stop the application with Ctrl+C.
```
s3_app_1 exited with code 0
Stopping localstack_main ... done
```