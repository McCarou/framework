# Example: simple SQS service

## Docker compose

### 1 Clone the repository

```shell
git clone https://github.com/radianteam/framework.git
cd framework
```

### 2 Goto this folder

```shell
cd example/sqs
```


### 3 Run the application

```shell
docker-compose up -d
```

### 4 Make a request to put a message to the input queue
Commands:
```shell
curl -X POST http://localhost:8080/ -H "Content-Type: application/text" -d "HellO!"
```

### 5 Make a request to get the message from the output queue
Commands:
```shell
curl -X GET http://localhost:8080/
```

Example:
```
curl -X GET http://localhost:8080/
HellO!
```

### 5 Enjoy!

And also don't forget to stop the application :)

```shell
docker-compose down
```
