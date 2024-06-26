package broadcast

import (
	"testing"
)

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
