package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	zmq "github.com/pebbe/zmq4"

	"github.com/thinkdoggie/zmq4chan/v2"
)

func main() {
	fmt.Println("Starting PUB/SUB example...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	// Start publisher
	wg.Add(1)
	go func() {
		defer wg.Done()
		runPublisher(ctx)
	}()

	// Start multiple subscribers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runSubscriber(ctx, id)
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Println("PUB/SUB example completed!")
}

func runPublisher(ctx context.Context) {
	// Create PUB socket
	pubSocket, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		log.Fatal("Failed to create PUB socket:", err)
	}
	defer pubSocket.Close()

	err = pubSocket.Bind("tcp://*:5556")
	if err != nil {
		log.Fatal("Failed to bind PUB socket:", err)
	}

	// Create channel adapter
	pubAdapter := zmq4chan.NewChanAdapter(pubSocket, 0, 100)
	defer pubAdapter.Close()

	pubAdapter.Start(ctx)

	fmt.Println("Publisher started on tcp://*:5556")

	// Wait for subscribers to connect
	time.Sleep(500 * time.Millisecond)

	// Publish messages with different topics
	topics := []string{"weather", "sports", "news"}
	messages := []string{
		"Sunny, 25°C",
		"Football match postponed",
		"Breaking: New discovery announced",
		"Rainy, 18°C",
		"Basketball finals tonight",
		"Stock market update",
	}

	for i, message := range messages {
		topic := topics[i%len(topics)]

		fmt.Printf("Publishing [%s]: %s\n", topic, message)

		// ZMQ PUB/SUB uses the first message part as the topic filter
		msg := [][]byte{[]byte(topic), []byte(message)}

		select {
		case pubAdapter.TxChan() <- msg:
			// Message sent successfully
		case <-ctx.Done():
			fmt.Println("Publisher context canceled")
			return
		}

		time.Sleep(1 * time.Second)
	}

	fmt.Println("Publisher finished sending messages")
}

func runSubscriber(ctx context.Context, id int) {
	// Create SUB socket
	subSocket, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		log.Printf("Subscriber %d: Failed to create SUB socket: %v", id, err)
		return
	}
	defer subSocket.Close()

	err = subSocket.Connect("tcp://localhost:5556")
	if err != nil {
		log.Printf("Subscriber %d: Failed to connect SUB socket: %v", id, err)
		return
	}

	// Each subscriber subscribes to different topics
	var subscription string
	switch id {
	case 0:
		subscription = "weather"
	case 1:
		subscription = "sports"
	case 2:
		subscription = "" // Subscribe to all messages
	}

	err = subSocket.SetSubscribe(subscription)
	if err != nil {
		log.Printf("Subscriber %d: Failed to set subscription: %v", id, err)
		return
	}

	// Create channel adapter
	subAdapter := zmq4chan.NewChanAdapter(subSocket, 100, 0)
	defer subAdapter.Close()

	subAdapter.Start(ctx)

	if subscription == "" {
		fmt.Printf("Subscriber %d: Listening for ALL messages\n", id)
	} else {
		fmt.Printf("Subscriber %d: Listening for '%s' messages\n", id, subscription)
	}

	// Receive messages
	for {
		select {
		case msg, ok := <-subAdapter.RxChan():
			if !ok {
				fmt.Printf("Subscriber %d: Receive channel closed\n", id)
				return
			}

			if len(msg) >= 2 {
				topic := string(msg[0])
				content := string(msg[1])
				fmt.Printf("Subscriber %d received [%s]: %s\n", id, topic, content)
			} else {
				fmt.Printf("Subscriber %d received malformed message: %v\n", id, msg)
			}

		case <-ctx.Done():
			fmt.Printf("Subscriber %d: Context canceled\n", id)
			return
		}
	}
}
