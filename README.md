# Client-Server Command Tool

## Building the Client CLI

To build the client CLI, move to the `client` directory and run:

```sh
cd client
go build -o ../build/cli client.go
```

## Server Setup

To run the server, navigate to the `server` directory and run:

```sh
go run server.go
```

## Running the Client CLI

To run the client CLI, use the following command:

```sh
cd build
./cli -p 4000 //Connect to port 4000
./cli --dump //List all UEs and gNodeB that server has

./cli --ue <node-name>
./cli --gnb <gnb-name>
```

For example:

```sh
./cli -ue imsi-001010000000001
```


