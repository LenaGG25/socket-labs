## Как запускать

Соберем исполняющие файлы:
```shell
go build server.go
go build client.go
```

Сначала запускаем сервер:
```shell
./server
```

Потом запускаем клиента:
```shell
./client <host> <port> <filename>
# example: 
```