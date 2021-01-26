# blog-grpc-gateway

This project is added here for Demo.

Enables protocol buffer compiler to read gRPC Service definitions and translate REST JSON API into gRPC.
Uses MongoDB.

Locally Start MongoDB or 
use Docker Compose
Example: [https://github.com/rsachdeva/illuminatingdeposits-grpc#start-mongodb](https://github.com/rsachdeva/illuminatingdeposits-grpc#start-mongodb)
```shell
sh ./runmongo.sh
```

Start gRPC Server
```shell
go run ./cmd/server
```

Start REST Server
```shell
go run ./cmd/restserver
```

Run gRPC Client to make gRPC Requests:
```shell
go run ./cmd/client
```

Run curl REST Client to make REST Http Requests:
```shell
sh ./runcurlrestclient.sh
```