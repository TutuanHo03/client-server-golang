# Client-Server Command Tool

## Building the Client CLI

To build the client CLI, navigate to the `client` directory and run:

```sh
go build -o ../build/cli client.go
```

## Running the Client CLI

To run the client CLI, use the following command:

```sh
./cli -ue <node-name>
```

For example:

```sh
./cli -ue imsi-001010000000001
```

## Server Setup

To run the server, navigate to the `server` directory and run:

```sh
go run server.go
```
