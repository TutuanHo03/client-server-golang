# Client-Server Command Tool

## Server Setup

To run the server, navigate to the `server` directory and run:

```sh
cd server
go build -o server
./server -c ../config/command.json
```

## Building the Client CLI

To build the client CLI, move to the `client` directory and run:

```sh
cd client
go build -o client

./client -p 4000 # Connect to port 4000 
./client --dump # List all UE and Gnb that server has
```

## Running the Client CLI

To run the client CLI, use the following command:

```sh
cd client
# For UE
./client -c ../config/commands.json -ue "<node-name>"
# For gNodeB
./client -c ../config/commands.json -gnb "<gnb-name>"
```

For example:

```sh
# For UE:
./client -c ../config/commands.json -ue "imsi-306956963543741"

# For gNodeB
./client -c ../config/commands.json -gnb "MSSIM-gnb-001-01-1"
```


