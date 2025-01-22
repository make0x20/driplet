# Driplet

![Driplet](assets/logo.png)

Driplet is a WebSocket microservice designed for real-time message broadcasting with targeting capabilities. It supports multiple endpoints, topic-based subscriptions, and message targeting based on client JWT claims.

DISCLAIMER: This is a new project and hasn't been battle-tested yet. Do your own testing before using it in production.

## Features

- Allows sending real-time messages of any kind from any backend service (specifically PHP applications that do not support WebSockets)
- Configurable multi-endpoints with separate JWT and API secrets for multiple applications
- Topic-based message subscriptions
- JWT authentication for WebSocket clients
- Advanced message targeting based on JWT claims
- HMAC signature validation for HTTP API
- Concurrent connection handling

## Use cases

- Push notifications
- Real-time page updates
- Real-time monitoring and alerting
- Chat applications
- Real-time updates on like/dislike, follow/unfollow, etc.
- Etc.

## How it works

![Driplet](assets/driplet.png)

## Install and run

```bash
go install github.com/make0x20/driplet@latest
driplet
```

Or build from source:

```bash
git clone https://github.com/make0x20/driplet
cd driplet
go build .
./driplet
```

### Run with Docker (easiest)

- Grab docker-compose.yml from the repository
- Run `docker-compose up -d`

### Run with Docker (manual)

```bash
git clone https://github.com/make0x20/driplet
cd driplet
docker build -t make0x20/driplet:latest .
docker-compose up -d
```

## Configuration

1. Create a configuration file `config.toml` (will generate a default one if not found):

```toml
[Global]
BindAddress = '0.0.0.0'
Port = 4719
LogFile = ''
LogLevel = 'normal'

[Endpoints]
[Endpoints.default]
Name = 'default'
APISecret = 'change-this-api-secret'
JWTSecret = 'change-this-jwt-secret'
```

2. Start the server:

```bash
./driplet
```

3. Connect to the WebSocket server:

```javascript
const ws = new WebSocket('ws://localhost:4719/default?token=your-jwt-token');
```

## Configuration

### Global

`BindAddress`: Server bind address (default: "0.0.0.0")

`Port`: Server port (default: 4719)

`LogFile`: Log file path (default: "")

`LogLevel`: Log level (default: "normal", options: "normal", "debug")

### Endpoints

Each endpoint requires:

`Name`: Endpoint identifier

`APISecret`: Secret for validating HTTP API requests

`JWTSecret`: Secret for validating client JWT tokens

All these values can be overridden by environment variables by prefixing them with `DRIPLET_` and converting them to uppercase.

For example, `BindAddress` can be overridden by setting `DRIPLET_BINDADDRESS`.

## WebSocket protocol

### Connection

Connect to the WebSocket server using the following URL scheme:

```
ws://server/ws/{endpoint}?token={jwt-token}
```

### Subscription to topics

Subscribe:

```json
{
  "type": "subscribe",
  "topic": "your-topic"
}
```

Unsubscribe:

```json
{
  "type": "unsubscribe",
  "topic": "your-topic"
}
```

## HTTP API

### Publish messages

POST `/api/{endpoint}/message`

Headers:
`X-Driplet-Signature`: HMAC signature of the request body

Body:

```json
{
  "message": {
    "your": "payload"
  },
  "topic": "target-topic",
  "target": {
    "include": {
      "role": "admin"
    },
    "exclude": {
      "uid": 5
    }
  }
}
```

## Message targeting

Messages can be targeted to specific clients based on their JWT claims:

`include`: Rules that clients must match (at least one)

`exclude`: Rules that exclude matching clients

## JWT example

Currently, Driplet matches claims based on the `custom` claim:

```json
{"typ":"JWT","alg":"HS256"}{"exp":1737564564,"custom":{"uid":"1","roles":["authenticated","administrator"]}}
```

Base64 encoded:

`eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJleHAiOjE3Mzc1NjQ1NjQsImN1c3RvbSI6eyJ1aWQiOiIxIiwicm9sZXMiOlsiYXV0aGVudGljYXRlZCIsImFkbWluaXN0cmF0b3IiXX19.VJ0jmWSntOimtv3KSA1I7lwkl9vFbmUCzvlN7j76fiQ`

