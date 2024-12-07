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
	// TODO: do not use maps, no thread safe for read and write
	messagesR1 map[int]*lockfree.Queue
	messagesR2 map[int]*lockfree.Queue
}

func (mq *MessageQueue) Enqueue(msg *Message) {
	if msg.r == 1 {
		mq.messagesR1[msg.s].Enqueue(msg)
	} else {
		mq.messagesR2[msg.s].Enqueue(msg)
	}

}

func (mq *MessageQueue) cleanOlds(r int, s int) {
	msgQueue := mq.messagesR1
	if r == 2 {
		msgQueue = mq.messagesR2
	}
	for i := range s {
		delete(msgQueue, i)
	}
}

const dequeueSleep = 50 * time.Millisecond

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
