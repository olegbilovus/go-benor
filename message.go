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
	messagesR1 map[int]*lockfree.Queue
	messagesR2 map[int]*lockfree.Queue
}

func (mq *MessageQueue) Queue(msg *Message) {
	if msg.r == 1 {
		mq.messagesR1[msg.s].Enqueue(msg)
	} else {
		mq.messagesR2[msg.s].Enqueue(msg)
	}

}

const dequeueSleep = 100 * time.Millisecond

func (mq *MessageQueue) Dequeue(r int, s int) *Message {
	var msg *Message
	msgQueue := mq.messagesR1
	if r == 2 {
		msgQueue = mq.messagesR2
	}
	for {
		o := msgQueue[s].Dequeue()
		if o != nil {
			msg = o.(*Message)
			break
		}
		if TERMINATE && decision.Load() != nil {
			msg = &Message{}
			break
		}
		time.Sleep(dequeueSleep)
	}
	return msg
}
