# zmq4chan - Go-native Channel interface for ZeroMQ sockets

[![CI](https://github.com/thinkdoggie/zmq4chan/workflows/CI/badge.svg)](https://github.com/thinkdoggie/zmq4chan/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/thinkdoggie/zmq4chan?status.svg)](https://godoc.org/github.com/thinkdoggie/zmq4chan)
[![License](https://img.shields.io/badge/License-BSD%202--Clause-orange.svg)](https://opensource.org/licenses/BSD-2-Clause)

A Go package that provides a channel-based adapter for [ZeroMQ](https://zeromq.org/) sockets, bridging ZMQ's message-passing model with Go's channel-based concurrency model.

## Features

- **Go Channel Interface**: Provide Go-native channels for ZMQ message sending and receiving
- **All ZMQ Socket Types**: Supports REQ/REP, DEALER/ROUTER, PUB/SUB, PUSH/PULL, PAIR, and more
- **Thread-Safe**: Concurrent access to channels is handled safely
- **Multi-part Messages**: Full support for ZMQ multi-part messages
- **Context Cancellation**: Graceful shutdown using Go's context pattern
- **Comprehensive Testing**: Comprehensive test coverage for various socket types and scenarios

## Installation

```bash
go get github.com/thinkdoggie/zmq4chan
```

## Prerequisites

This package depends on [zmq4](https://github.com/pebbe/zmq4), which requires ZeroMQ 4.x to be installed on your system.

### Installing ZeroMQ

**macOS:**
```bash
brew install zeromq
```

**Ubuntu/Debian:**
```bash
sudo apt-get install libzmq3-dev
```

**RHEL/CentOS:**
```bash
sudo yum install zeromq-devel
```

## Quick Start

Here's a simple REQ/REP example:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    zmq "github.com/pebbe/zmq4"
    "github.com/thinkdoggie/zmq4chan"
)

func main() {
    // Create REP socket (server)
    repSocket, err := zmq.NewSocket(zmq.REP)
    if err != nil {
        log.Fatal(err)
    }
    defer repSocket.Close()
    
    err = repSocket.Bind("tcp://*:5555")
    if err != nil {
        log.Fatal(err)
    }

    // Create channel adapter for REP socket
    repAdapter := zmq4chan.NewChanAdapter(repSocket, 10, 10)
    defer repAdapter.Close()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    repAdapter.Start(ctx)

    // Server goroutine
    go func() {
        for {
            select {
            case msg := <-repAdapter.RxChan():
                fmt.Printf("Server received: %s\n", string(msg[0]))
                // Echo back the message
                reply := [][]byte{[]byte("Echo: " + string(msg[0]))}
                repAdapter.TxChan() <- reply
            case <-ctx.Done():
                return
            }
        }
    }()

    // Create REQ socket (client)
    reqSocket, err := zmq.NewSocket(zmq.REQ)
    if err != nil {
        log.Fatal(err)
    }
    defer reqSocket.Close()
    
    err = reqSocket.Connect("tcp://localhost:5555")
    if err != nil {
        log.Fatal(err)
    }

    reqAdapter := zmq4chan.NewChanAdapter(reqSocket, 10, 10)
    defer reqAdapter.Close()
    
    reqAdapter.Start(ctx)

    // Send a request
    request := [][]byte{[]byte("Hello, World!")}
    reqAdapter.TxChan() <- request

    // Receive reply
    select {
    case reply := <-reqAdapter.RxChan():
        fmt.Printf("Client received: %s\n", string(reply[0]))
    case <-time.After(5 * time.Second):
        fmt.Println("Request timeout")
    }
}
```

## API Reference

For complete API documentation, visit [pkg.go.dev/github.com/thinkdoggie/zmq4chan](https://pkg.go.dev/github.com/thinkdoggie/zmq4chan).

### Core Types

#### `ChanAdapter`
The main type that bridges ZMQ sockets with Go channels.

```go
type ChanAdapter struct {
    // ... internal fields
}
```

### Key Functions

| Function | Description |
|----------|-------------|
| `NewChanAdapter(socket, rxSize, txSize)` | Creates a new adapter for a ZMQ socket |
| `Start(ctx)` | Starts the adapter with context for cancellation |
| `RxChan()` | Returns receive-only channel for incoming messages |
| `TxChan()` | Returns send-only channel for outgoing messages |
| `Close()` | Gracefully shuts down the adapter |

### Message Format

Messages are represented as `[][]byte` where each `[]byte` is a message part:
- **Single-part**: `[][]byte{[]byte("hello")}`
- **Multi-part**: `[][]byte{[]byte("part1"), []byte("part2")}`


## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -am 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the BSD 2-Clause License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built on top of the excellent [zmq4](https://github.com/pebbe/zmq4) Go bindings
- Inspired by the need to integrate ZMQ with idiomatic Go patterns 