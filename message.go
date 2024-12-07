package main

import (
	"fmt"
	"golang.design/x/lockfree"
	"time"
)

type Message struct {
	r int
	s int
	v V
	p int
}

func (m *Message) String() string {
	return fmt.Sprintf("(r:%v, s:%v, v:%v)", m.r, m.s, m.v)
}

type MessageQueue struct {
	messages *lockfree.Queue
}

func (mq *MessageQueue) Queue(msg *Message) {
	mq.messages.Enqueue(msg)

}

const popSleep = 50 * time.Millisecond

func (mq *MessageQueue) Dequeue() *Message {
	var msg *Message
	for {
		o := mq.messages.Dequeue()
		if o != nil {
			msg = o.(*Message)
			break
		}
		time.Sleep(popSleep)
	}
	return msg
}
