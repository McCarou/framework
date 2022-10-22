# Example application for the Radian framework

## Usage

#### 1 Clone the repository

```
git clone https://github.com/radianteam/framework.git
```
```
cd framework
```

#### 2 Goto this folder

```
cd example/rest_simple
```


#### 3 Run the application

```
docker-compose up -d
```

#### 4 Make a request
Commands:
```
curl 127.0.0.1:8088/ 
```

Example
```
radian@radian:~$ curl 127.0.0.1:8088/                                   
Hello world!
```

#### 5 Enjoy!

And don't forget to stop the application :)

```
docker-compose down
```
