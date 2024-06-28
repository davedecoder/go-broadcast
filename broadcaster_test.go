package broadcast

import (
	"testing"
	"time"
)

func TestNewHUB(t *testing.T) {
	bufflen := 10
	h := NewHUB(bufflen)
	if h == nil {
		t.Error("NewHUB returned nil")
	}
	if len(h.input) != 0 || cap(h.input) != bufflen {
		t.Errorf("HUB input channel buffer length expected to be %d, got %d", bufflen, cap(h.input))
	}
}

func TestSubscribeAndUnsubscribe(t *testing.T) {
	h := NewHUB(10)
	defer h.Close()
	subscriber := NewSubscriber()
	h.Subscribe(subscriber)
	if _, ok := h.subscribers[subscriber]; !ok {
		t.Error("Subscriber was not added")
	}

	h.Unsubscribe(subscriber)

	select {
	case <-subscriber.closeChan:
		time.Sleep(time.Second * 1)
		// Broadcast goroutine finished successfully
	case <-time.After(time.Second * 1):
		t.Error("Timed out waiting for broadcast goroutine to finish")
	}
	if _, ok := h.subscribers[subscriber]; ok {
		t.Error("Subscriber was not removed")
	}
}

func TestSubscribeAndUnsubscribe2(t *testing.T) {
	h := NewHUB(10)
	defer h.Close()

	subscriber := NewSubscriber()
	h.Subscribe(subscriber)

	h.Publish("Hello, World!")

	select {
	case msg := <-subscriber.MsgChan:
		if msg != "Hello, World!" {
			t.Errorf("Received incorrect message: %v", msg)
		}
	case <-time.After(time.Second * 1):
		t.Error("Timed out waiting for message")
	}

	h.Unsubscribe(subscriber)
	h.Publish("Hola Mundo!")

	select {
	case <-subscriber.MsgChan:
		t.Error("Received unexpected message")
	case <-time.After(time.Second * 1):
		// Expected behavior
	}
}

func TestHUBClose(t *testing.T) {
	h := NewHUB(10)
	subscriber := NewSubscriber()
	h.Subscribe(subscriber)

	h.Close()

	select {
	case <-h.done:
		// Expected behavior
	case <-time.After(time.Second * 1):
		t.Error("Timed out waiting for broadcast goroutine to finish")
	}
}

func TestHUB_broadcast(t *testing.T) {
	// Create a new HUB instance
	hub := &HUB{
		input:       make(chan interface{}),
		reg:         make(chan *Subscriber),
		unreg:       make(chan *Subscriber),
		subscribers: make(map[*Subscriber]struct{}),
	}

	// Create some test subscribers
	sub1 := &Subscriber{
		MsgChan:   make(chan interface{}, 1),
		closeChan: make(chan struct{}),
	}
	sub2 := &Subscriber{
		MsgChan:   make(chan interface{}, 1),
		closeChan: make(chan struct{}),
	}

	// Add the test subscribers to the HUB
	hub.subscribers[sub1] = struct{}{}
	hub.subscribers[sub2] = struct{}{}

	// Test case 1: Broadcast a message to all subscribers
	testMsg := "Hello, World!"
	hub.broadcast(testMsg)

	// Check if the message was received by both subscribers
	select {
	case msg := <-sub1.MsgChan:
		if msg != testMsg {
			t.Errorf("Subscriber 1 received incorrect message: %v", msg)
		}
	case <-sub1.closeChan:
		t.Errorf("Subscriber 1 closed unexpectedly")
	default:
		t.Errorf("Subscriber 1 did not receive the message")
	}

	select {
	case msg := <-sub2.MsgChan:
		if msg != testMsg {
			t.Errorf("Subscriber 2 received incorrect message: %v", msg)
		}
	case <-sub2.closeChan:
		t.Errorf("Subscriber 2 closed unexpectedly")
	default:
		t.Errorf("Subscriber 2 did not receive the message")
	}

	// Test case 2: Broadcast a message to a closed subscriber
	closedSub := &Subscriber{
		MsgChan:   make(chan interface{}),
		closeChan: make(chan struct{}),
	}
	close(closedSub.closeChan)
	hub.subscribers[closedSub] = struct{}{}

	hub.broadcast("Test message")

	// Check if the message was not sent to the closed subscriber
	select {
	case <-closedSub.MsgChan:
		t.Errorf("Message was sent to a closed subscriber")
	default:
		// Expected behavior
	}
}

func TestPublish(t *testing.T) {
	h := NewHUB(10)
	msg := "test message"
	subscriber := NewSubscriber()
	h.Subscribe(subscriber)

	go h.Publish(msg)
	select {
	case m := <-subscriber.MsgChan:
		if m != msg {
			t.Errorf("Expected message %v, got %v", msg, m)
		}
	case <-time.After(1 * time.Second):
		t.Error("Message was not received within 1 second")
	}
}

func TestSubscriber_Receive(t *testing.T) {
	// Create a new Subscriber instance
	sub := &Subscriber{
		MsgChan:   make(chan interface{}, 1),
		closeChan: make(chan struct{}),
	}

	// Test case 1: Receive a message from the channel
	testMsg := "Hello, World!"
	sub.MsgChan <- testMsg

	msg := sub.Receive()
	if msg != testMsg {
		t.Errorf("Received incorrect message: %v", msg)
	}

	// Test case 2: Receive from a closed channel
	close(sub.MsgChan)

	msg, ok := <-sub.MsgChan
	if ok {
		t.Errorf("Received unexpected message from closed channel: %v", msg)
	}
}

func BenchmarkDirectSend(b *testing.B) {
	h := NewHUB(10)
	defer h.Close()
	subscriber := NewSubscriber()
	h.Subscribe(subscriber)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Publish(i)
		subscriber.Receive()
	}
}

func BenchmarkParallenDirectSend(b *testing.B) {
	h := NewHUB(10)
	defer h.Close()

	subscriber := NewSubscriber()
	h.Subscribe(subscriber)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Publish(1)
			subscriber.Receive()
		}
	})
}
