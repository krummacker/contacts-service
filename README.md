# Contacts-Service

This repository contains the code for a simple contacts REST service, implemented in Golang.

All shell commands below assume that you are in the project root.

## How to run unit tests

This does not require running MySQL and REST services.

```bash
go test -v $(go list ./... | grep -v integrationtest)
```

## How to run integration tests

Make sure that MySQL is running locally. This does not require running REST services.

```bash
DBHOST=localhost DBUSER=<local user> DBPWD=<password> go test -v ./internal/integrationtest
```

## How to run manual tests

Make sure that MySQL is running locally.

In one shell, start the service with verbose logging turned on:

```bash
PORT=8080 DBHOST=localhost DBUSER=<local user> DBPWD=<password> go run cmd/service/main.go
```

In a second shell, call the REST URLs, for example:

```bash
curl "http://localhost:8080/contacts?firstname=Ivan&lastname=Gentry"
curl "http://localhost:8080/contacts?orderby=firstname&ascending=false"
```

## How to run performance tests

Make sure that MySQL is running locally.

In one shell, start the service with verbose logging turned off:

```bash
PORT=8080 DBHOST=localhost DBUSER=<local user> DBPWD=<password> GIN_MODE=release GIN_LOGGING=OFF go run cmd/service/main.go
```

In a second shell, run the tests:

```bash
PORT=8080 go run cmd/perftest/main.go
```

## How to build and run with Docker

Make sure that Docker and MySQL are running locally. Build the image from the project root:

```bash
docker build -t contacts-service:latest .
```

Run the service and publish its port to the host. When MySQL is running on the host, use `host.docker.internal` as the database host:

```bash
docker run --rm --name contacts-service -p 8080:8080 \
	-e PORT=8080 \
	-e DBHOST=host.docker.internal \
	-e DBUSER=<local user> \
	-e DBPWD=<password> \
	-e GIN_MODE=release \
	-e GIN_LOGGING=OFF \
	contacts-service:latest
```

The service is then available at `http://localhost:8080`.
