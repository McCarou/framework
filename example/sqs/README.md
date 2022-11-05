# Example: simple REST service

1. [`Manual`](#1-manual) 
2. [`Docker compose`](#2-docker-compose)

## 1 Manual

TODO

## 2 Docker compose

### 1 Clone the repository

```
git clone https://github.com/radianteam/framework.git
```
```
cd framework
```

### 2 Goto this folder

```
cd example/sqs
```


### 3 Run the application

```
docker-compose up -d
```

### 4 Make a request to put a message to the input queue
Commands:
```
curl -X POST http://localhost:8080/ -H "Content-Type: application/text" -d "HellO!"
```

### 5 Make a request to get the message from the output queue
Commands:
```
curl -X GET http://localhost:8080/
```

Example:
```
curl -X GET http://localhost:8080/
HellO!
```

### 5 Enjoy!

And don't forget to stop the application :)

```
docker-compose down
```
