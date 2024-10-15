# Scalable Game Server

A service for on-demand spawning multiple instances of single-lobby game servers.

## Requirements

* Docker
* Golang

## Installation

### Building From Source
```
git clone https://github.com/rkblake/scalable-game-server.git
cd scalabe-game-server
go build .
```

### Building Docker Image
```
# with the game server in the docker folder named "server"
cd docker
docker build . -t server:latest
```

### Starting the Server
```
./game-server
```

