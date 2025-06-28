package zmq4chan

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	zmq "github.com/pebbe/zmq4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ChanAdapterTestFixture struct {
	// Begin of test case parameters
	socketAddr string
	rxChanSize int
	txChanSize int
	identity   []byte
	// End of test case parameters
	bindSocket     *zmq.Socket
	bindAdapter    *ChanAdapter
	cancel         func()
	connectSocket  *zmq.Socket
	connectAdapter *ChanAdapter
}

func (tf *ChanAdapterTestFixture) setUp(t *testing.T, connectType, bindType zmq.Type) {
	var err error
	tf.bindSocket, err = zmq.NewSocket(bindType)
	require.NoError(t, err)

	err = tf.bindSocket.Bind(tf.socketAddr)
	require.NoError(t, err)

	tf.bindAdapter = NewChanAdapter(tf.bindSocket, tf.rxChanSize, tf.txChanSize)

	tf.connectSocket, err = zmq.NewSocket(connectType)
	require.NoError(t, err)

	if tf.identity != nil {
		err = tf.connectSocket.SetIdentity(string(tf.identity[:]))
		require.NoError(t, err)
	}

	err = tf.connectSocket.Connect(tf.socketAddr)
	require.NoError(t, err)

	tf.connectAdapter = NewChanAdapter(tf.connectSocket, tf.rxChanSize, tf.txChanSize)

	ctx, cancel := context.WithCancel(context.Background())
	tf.cancel = cancel
	tf.bindAdapter.Start(ctx)
	tf.connectAdapter.Start(ctx)

	t.Cleanup(func() {
		if tf.cancel != nil {
			tf.cancel()
		}
		if tf.bindAdapter != nil {
			tf.bindAdapter.Close()
		}
		if tf.connectAdapter != nil {
			tf.connectAdapter.Close()
		}
		if tf.bindSocket != nil {
			tf.bindSocket.SetLinger(0)
			tf.bindSocket.Close()
		}
		if tf.connectSocket != nil {
			tf.connectSocket.SetLinger(0)
			tf.connectSocket.Close()
		}
	})
}

// Test cases
func TestChanAdapterReq2Rep(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tf.setUp(t, zmq.REQ, zmq.REP)

	// Test
	requestMsg := [][]byte{[]byte("Hello, world!")}
	replyMsg := [][]byte{[]byte("Ack")}

	tf.connectAdapter.TxChan() <- requestMsg
	gotBindMsg, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "bindAdapter.RxChan() should return a message")
	assert.Equal(t, requestMsg, gotBindMsg, "bindAdapter received wrong message")

	tf.bindAdapter.TxChan() <- replyMsg
	gotConnectMsg, ok := <-tf.connectAdapter.RxChan()
	require.True(t, ok, "connectAdapter.RxChan() should return a message")
	assert.Equal(t, replyMsg, gotConnectMsg, "connectAdapter received wrong message")
}

func TestChanAdapterReq2Router(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
		identity:   []byte("client-12345"),
	}
	tf.setUp(t, zmq.REQ, zmq.ROUTER)

	// Test
	emptyDelimiter := []byte{}
	routeHeader := [][]byte{tf.identity, emptyDelimiter}
	replyMsg := [][]byte{[]byte("Reply")}
	requestMsg := [][]byte{[]byte("Request"), []byte("Hello, world!")}
	recvRequestMsg := append(routeHeader, requestMsg...)
	sendReplyMsg := append(routeHeader, replyMsg...)

	tf.connectAdapter.TxChan() <- requestMsg
	gotBindMsg, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "bindAdapter.RxChan() should return a message")
	assert.Equal(t, recvRequestMsg, gotBindMsg, "bindAdapter received wrong message")

	tf.bindAdapter.TxChan() <- sendReplyMsg
	gotConnectMsg, ok := <-tf.connectAdapter.RxChan()
	require.True(t, ok, "connectAdapter.RxChan() should return a message")
	assert.Equal(t, replyMsg, gotConnectMsg, "connectAdapter received wrong message")
}

func TestChanAdapterDealer2Router(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
		identity:   []byte("client-12345"),
	}
	tf.setUp(t, zmq.DEALER, zmq.ROUTER)

	// Test
	emptyDelimiter := []byte{}
	routeHeader := [][]byte{tf.identity, emptyDelimiter}
	dealerHeader := [][]byte{emptyDelimiter}
	replyMsg := [][]byte{[]byte("Reqly")}
	requestMsg := [][]byte{[]byte("Request"), []byte("Hello, world!")}
	sendRequestMsg := append(dealerHeader, requestMsg...)
	recvRequestMsg := append(routeHeader, requestMsg...)
	sendReplyMsg := append(routeHeader, replyMsg...)
	recvReplyMsg := append(dealerHeader, replyMsg...)

	tf.connectAdapter.TxChan() <- sendRequestMsg
	gotBindMsg, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "bindAdapter.RxChan() should return a message")
	assert.Equal(t, recvRequestMsg, gotBindMsg, "bindAdapter received wrong message")

	tf.bindAdapter.TxChan() <- sendReplyMsg
	gotConnectMsg, ok := <-tf.connectAdapter.RxChan()
	require.True(t, ok, "connectAdapter.RxChan() should return a message")
	assert.Equal(t, recvReplyMsg, gotConnectMsg, "connectAdapter received wrong message")
}

func TestChanAdapterReconnectDealer2Router(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
		identity:   []byte("client-12345"),
	}
	tf.setUp(t, zmq.DEALER, zmq.ROUTER)

	// Test
	emptyDelimiter := []byte{}
	routeHeader := [][]byte{tf.identity, emptyDelimiter}
	dealerHeader := [][]byte{emptyDelimiter}
	replyMsg := [][]byte{[]byte("Reqly")}
	requestMsg := [][]byte{[]byte("Request"), []byte("Hello, world!")}
	sendDropMsg := append(dealerHeader, replyMsg...)
	sendRequestMsg := append(dealerHeader, requestMsg...)
	recvRequestMsg := append(routeHeader, requestMsg...)
	sendReplyMsg := append(routeHeader, replyMsg...)
	recvReplyMsg := append(dealerHeader, replyMsg...)

	tf.connectAdapter.TxChan() <- sendDropMsg
	tf.connectAdapter.Close()
	tf.connectSocket.SetLinger(0)
	tf.connectSocket.Close()

	time.Sleep(100 * time.Millisecond)
	rxChan := tf.bindAdapter.RxChan()
	// drop all messages in the channel
drainLoop:
	for {
		select {
		case _, ok := <-rxChan:
			if !ok {
				break drainLoop
			}
		default:
			break drainLoop
		}
	}

	var err error
	tf.connectSocket, err = zmq.NewSocket(zmq.DEALER)
	require.NoError(t, err)
	err = tf.connectSocket.SetIdentity(string(tf.identity))
	require.NoError(t, err)
	err = tf.connectSocket.Connect(tf.socketAddr)
	require.NoError(t, err)
	tf.connectAdapter = NewChanAdapter(tf.connectSocket, tf.rxChanSize, tf.txChanSize)
	tf.connectAdapter.Start(context.Background())
	defer tf.connectAdapter.Close()

	tf.connectAdapter.TxChan() <- sendRequestMsg
	gotBindMsg, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "bindAdapter.RxChan() should return a message")
	assert.Equal(t, recvRequestMsg, gotBindMsg, "bindAdapter received wrong message")

	tf.bindAdapter.TxChan() <- sendReplyMsg
	gotConnectMsg, ok := <-tf.connectAdapter.RxChan()
	require.True(t, ok, "connectAdapter.RxChan() should return a message")
	assert.Equal(t, recvReplyMsg, gotConnectMsg, "connectAdapter received wrong message")
}

func TestChanAdapterPub2Sub(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tf.setUp(t, zmq.SUB, zmq.PUB)

	// Subscribe to all messages
	err := tf.connectSocket.SetSubscribe("")
	require.NoError(t, err)

	// Give subscription time to propagate
	time.Sleep(100 * time.Millisecond)

	// Test
	pubMsg := [][]byte{[]byte("published message")}

	tf.bindAdapter.TxChan() <- pubMsg
	gotConnectMsg, ok := <-tf.connectAdapter.RxChan()
	require.True(t, ok, "connectAdapter.RxChan() should return a message")
	assert.Equal(t, pubMsg, gotConnectMsg, "connectAdapter received wrong message")
}

func TestChanAdapterPush2Pull(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tf.setUp(t, zmq.PUSH, zmq.PULL)

	// Test
	pushMsg := [][]byte{[]byte("pushed message")}

	tf.connectAdapter.TxChan() <- pushMsg
	gotBindMsg, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "bindAdapter.RxChan() should return a message")
	assert.Equal(t, pushMsg, gotBindMsg, "bindAdapter received wrong message")
}

func TestChanAdapterPair2Pair(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tf.setUp(t, zmq.PAIR, zmq.PAIR)

	// Test bidirectional communication
	msg1 := [][]byte{[]byte("message from connect to bind")}
	msg2 := [][]byte{[]byte("message from bind to connect")}

	tf.connectAdapter.TxChan() <- msg1
	gotBindMsg, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "bindAdapter.RxChan() should return a message")
	assert.Equal(t, msg1, gotBindMsg, "bindAdapter received wrong message")

	tf.bindAdapter.TxChan() <- msg2
	gotConnectMsg, ok := <-tf.connectAdapter.RxChan()
	require.True(t, ok, "connectAdapter.RxChan() should return a message")
	assert.Equal(t, msg2, gotConnectMsg, "connectAdapter received wrong message")
}

func TestChanAdapterEmptyMessage(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tf.setUp(t, zmq.PUSH, zmq.PULL)

	// Test sending an empty message
	emptyMsg := [][]byte{[]byte{}}

	tf.connectAdapter.TxChan() <- emptyMsg
	gotBindMsg, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "bindAdapter.RxChan() should return a message")
	assert.Equal(t, emptyMsg, gotBindMsg, "bindAdapter received wrong message")
}

func TestChanAdapterLargeMessage(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tf.setUp(t, zmq.PUSH, zmq.PULL)

	// Test sending a large message (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	largeMsg := [][]byte{largeData}

	tf.connectAdapter.TxChan() <- largeMsg
	gotBindMsg, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "bindAdapter.RxChan() should return a message")
	assert.Equal(t, largeMsg, gotBindMsg, "bindAdapter received wrong message")
}

func TestChanAdapterMultipartMessages(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tf.setUp(t, zmq.PUSH, zmq.PULL)

	// Test multipart message
	multipartMsg := [][]byte{
		[]byte("part1"),
		[]byte("part2"),
		[]byte("part3"),
		[]byte("part4"),
	}

	tf.connectAdapter.TxChan() <- multipartMsg
	gotBindMsg, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "bindAdapter.RxChan() should return a message")
	assert.Equal(t, multipartMsg, gotBindMsg, "bindAdapter received wrong message")
}

func TestChanAdapterCloseBeforeStart(t *testing.T) {
	socket, err := zmq.NewSocket(zmq.REP)
	require.NoError(t, err)
	defer socket.Close()

	adapter := NewChanAdapter(socket, 10, 10)

	// Close before starting - should not panic
	assert.NotPanics(t, func() {
		adapter.Close()
	})
}

func TestChanAdapterChannelSizes(t *testing.T) {
	testCases := []struct {
		name       string
		rxChanSize int
		txChanSize int
	}{
		{"rx:0,tx:0", 0, 0},
		{"rx:1,tx:1", 1, 1},
		{"rx:100,tx:50", 100, 50},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tf := ChanAdapterTestFixture{
				socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
				rxChanSize: testCase.rxChanSize,
				txChanSize: testCase.txChanSize,
			}
			tf.setUp(t, zmq.PUSH, zmq.PULL)

			// Test basic functionality with different channel sizes
			testMsg := [][]byte{[]byte("test message")}

			tf.connectAdapter.TxChan() <- testMsg
			gotBindMsg, ok := <-tf.bindAdapter.RxChan()
			require.True(t, ok, "bindAdapter.RxChan() should return a message")
			assert.Equal(t, testMsg, gotBindMsg, "bindAdapter received wrong message")
		})
	}
}

func TestChanAdapterConcurrentAccess(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 100,
		txChanSize: 100,
	}
	tf.setUp(t, zmq.DEALER, zmq.ROUTER)

	const numGoroutines = 10
	const messagesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // senders + receivers

	// Multiple goroutines sending messages
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := [][]byte{[]byte{}, []byte(fmt.Sprintf("msg-%d-%d", id, j))}
				tf.connectAdapter.TxChan() <- msg
			}
		}(i)
	}

	// Multiple goroutines receiving messages
	received := make(chan [][]byte, numGoroutines*messagesPerGoroutine)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				select {
				case msg := <-tf.bindAdapter.RxChan():
					received <- msg
				case <-time.After(5 * time.Second):
					assert.Fail(t, "Timeout waiting for message")
					return
				}
			}
		}()
	}

	wg.Wait()
	close(received)

	// Verify we received all messages
	count := 0
	for range received {
		count++
	}

	assert.Equal(t, numGoroutines*messagesPerGoroutine, count, "Expected all messages to be received")
}

func TestChanAdapterHighVolumeMessages(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 1000,
		txChanSize: 1000,
	}
	tf.setUp(t, zmq.PUSH, zmq.PULL)

	const numMessages = 1000
	done := make(chan bool)

	// Receiver goroutine
	go func() {
		for i := 0; i < numMessages; i++ {
			select {
			case msg := <-tf.bindAdapter.RxChan():
				expected := fmt.Sprintf("message-%d", i)
				assert.Equal(t, expected, string(msg[0]), "Message content mismatch")
			case <-time.After(5 * time.Second):
				assert.Fail(t, fmt.Sprintf("Timeout waiting for message %d", i))
				return
			}
		}
		done <- true
	}()

	// Send messages rapidly
	for i := 0; i < numMessages; i++ {
		msg := [][]byte{[]byte(fmt.Sprintf("message-%d", i))}
		tf.connectAdapter.TxChan() <- msg
	}

	// Wait for completion
	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		assert.Fail(t, "Test timed out")
	}
}

func TestChanAdapterBackpressure(t *testing.T) {
	// Small buffer to test backpressure
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 2, // Very small buffer
		txChanSize: 2,
	}
	// Use PUSH-PULL instead of REQ-REP to avoid strict alternating requirements
	tf.setUp(t, zmq.PUSH, zmq.PULL)

	// Send multiple messages without reading them
	// This should test backpressure handling
	sent := 0
sendLoop:
	for i := 0; i < 5; i++ {
		msg := [][]byte{[]byte(fmt.Sprintf("msg-%d", i))}
		select {
		case tf.connectAdapter.TxChan() <- msg:
			sent++
		case <-time.After(100 * time.Millisecond):
			// This is expected due to backpressure
			break sendLoop
		}
	}

	// Now start reading messages
	for i := 0; i < sent; i++ {
		select {
		case <-tf.bindAdapter.RxChan():
			// Good, received a message
		case <-time.After(1 * time.Second):
			assert.Fail(t, fmt.Sprintf("Timeout receiving message %d", i))
		}
	}
}

func TestChanAdapterErrorRecovery(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tf.setUp(t, zmq.REQ, zmq.REP)

	// Send a normal message first
	normalMsg := [][]byte{[]byte("normal")}
	tf.connectAdapter.TxChan() <- normalMsg

	_, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "Should receive normal message")

	// Force close the underlying socket to simulate an error
	tf.connectSocket.Close()

	// Give some time for the error to propagate
	time.Sleep(100 * time.Millisecond)

	// The adapter should handle this gracefully with no panic or deadlock
}

func TestChanAdapterMessageOrder(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 100,
		txChanSize: 100,
	}
	tf.setUp(t, zmq.PUSH, zmq.PULL)

	const numMessages = 20

	// Send messages in order
	for i := 0; i < numMessages; i++ {
		msg := [][]byte{[]byte(fmt.Sprintf("msg-%d", i))}
		tf.connectAdapter.TxChan() <- msg
	}

	// Verify they arrive in order
	for i := 0; i < numMessages; i++ {
		select {
		case msg := <-tf.bindAdapter.RxChan():
			expected := fmt.Sprintf("msg-%d", i)
			assert.Equal(t, expected, string(msg[0]), "Messages should arrive in order")
		case <-time.After(1 * time.Second):
			assert.Fail(t, fmt.Sprintf("Timeout waiting for message %d", i))
		}
	}
}

func TestChanAdapterZeroByteMessage(t *testing.T) {
	tf := ChanAdapterTestFixture{
		socketAddr: fmt.Sprintf("inproc://test-%d", time.Now().UnixNano()),
		rxChanSize: 10,
		txChanSize: 10,
	}
	tf.setUp(t, zmq.PUSH, zmq.PULL)

	// Test message with zero-length byte array
	zeroMsg := [][]byte{make([]byte, 0)}

	tf.connectAdapter.TxChan() <- zeroMsg
	gotBindMsg, ok := <-tf.bindAdapter.RxChan()
	require.True(t, ok, "bindAdapter.RxChan() should return a message")
	assert.Equal(t, zeroMsg, gotBindMsg, "bindAdapter received wrong message")
}
