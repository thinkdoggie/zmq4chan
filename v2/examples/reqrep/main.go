package main

import (
	"context"
	"fmt"
	"log"
	"time"

	zmq "github.com/pebbe/zmq4"

	"github.com/thinkdoggie/zmq4chan/v2"
)

func main() {
	fmt.Println("Starting REQ/REP example...")

	// Create REP socket (server)
	repSocket, err := zmq.NewSocket(zmq.REP)
	if err != nil {
		log.Fatal("Failed to create REP socket:", err)
	}
	defer repSocket.Close()

	err = repSocket.Bind("tcp://*:5555")
	if err != nil {
		log.Fatal("Failed to bind REP socket:", err)
	}

	// Create channel adapter for REP socket
	repAdapter, err := zmq4chan.NewChanAdapter(repSocket, 10, 10)
	if err != nil {
		log.Fatal("Failed to create REP adapter:", err)
	}
	defer repAdapter.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repAdapter.Start(ctx)

	// Server goroutine
	go func() {
		fmt.Println("Server listening on tcp://*:5555")
		for {
			select {
			case msg, ok := <-repAdapter.RxChan():
				if !ok {
					fmt.Println("Server receive channel closed")
					return
				}
				fmt.Printf("Server received: %s\n", string(msg[0]))
				// Echo back the message with a prefix
				reply := [][]byte{[]byte("Echo: " + string(msg[0]))}
				repAdapter.TxChan() <- reply
			case <-ctx.Done():
				fmt.Println("Server context canceled")
				return
			}
		}
	}()

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)

	// Create REQ socket (client)
	reqSocket, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		log.Fatal("Failed to create REQ socket:", err)
	}
	defer reqSocket.Close()

	err = reqSocket.Connect("tcp://localhost:5555")
	if err != nil {
		log.Fatal("Failed to connect REQ socket:", err)
	}

	reqAdapter, err := zmq4chan.NewChanAdapter(reqSocket, 10, 10)
	if err != nil {
		log.Fatal("Failed to create REQ adapter:", err)
	}
	defer reqAdapter.Close()

	reqAdapter.Start(ctx)

	// Send multiple requests
	messages := []string{"Hello, World!", "How are you?", "Goodbye!"}

	for _, msgText := range messages {
		fmt.Printf("Client sending: %s\n", msgText)

		// Send request
		request := [][]byte{[]byte(msgText)}
		reqAdapter.TxChan() <- request

		// Receive reply
		select {
		case reply, ok := <-reqAdapter.RxChan():
			if !ok {
				fmt.Println("Client receive channel closed")
				return
			}
			fmt.Printf("Client received: %s\n", string(reply[0]))
		case <-time.After(2 * time.Second):
			fmt.Println("Request timeout")
			return
		case <-ctx.Done():
			fmt.Println("Context canceled")
			return
		}

		// Small delay between requests
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("Example completed successfully!")
}
