/*
Package broadcast provides pubsub of messages over channels.

A provider has a Broadcaster into which it Submits messages and into
which subscribers Register to pick up those messages.
*/
package broadcast

type Subscriber struct {
	MsgChan   chan interface{}
	closeChan chan struct{}
}

func NewSubscriber() *Subscriber {
	return &Subscriber{
		MsgChan:   make(chan interface{}),
		closeChan: make(chan struct{}),
	}
}

func (s *Subscriber) Receive() interface{} {
	return <-s.MsgChan
}

type HUB struct {
	input chan interface{}
	reg   chan *Subscriber
	unreg chan *Subscriber

	subscribers map[*Subscriber]struct{}
	done        chan struct{}
}

func NewHUB(bufflen int) *HUB {
	h := &HUB{
		input:       make(chan interface{}, bufflen),
		reg:         make(chan *Subscriber),
		unreg:       make(chan *Subscriber),
		subscribers: make(map[*Subscriber]struct{}),
		done:        make(chan struct{}),
	}

	go h.run()

	return h
}

func (h *HUB) Subscribe(subscriber *Subscriber) {
	h.reg <- subscriber
}

func (h *HUB) Unsubscribe(subscriber *Subscriber) {
	close(subscriber.closeChan)
	h.unreg <- subscriber
}

func (h *HUB) Publish(msg interface{}) {
	h.input <- msg
}

func (h *HUB) Close() {
	close(h.reg)
	close(h.unreg)
	close(h.input)
	close(h.done)
}

func (h *HUB) broadcast(msg interface{}) {
	for s := range h.subscribers {
		select {
		case s.MsgChan <- msg:
		case <-s.closeChan:
		}
	}
}

func (h *HUB) run() {
	for {
		select {
		case ch, ok := <-h.reg:
			if ok {
				h.subscribers[ch] = struct{}{}
			} else {
				return
			}
		case ch := <-h.unreg:
			delete(h.subscribers, ch)
		case msg := <-h.input:
			h.broadcast(msg)
		case <-h.done:
			return
		}
	}
}
