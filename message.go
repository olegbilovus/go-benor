package main

import (
	"fmt"
	"sync"
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
	messagesR1 map[int][]*Message
	messagesR2 map[int][]*Message

	muR1 *sync.Mutex
	muR2 *sync.Mutex

	notEmptyR1 map[int]*sync.Cond
	notEmptyR2 map[int]*sync.Cond
}

func (mq *MessageQueue) Enqueue(msg *Message) {
	r, s := msg.r, msg.s

	msgs := mq.messagesR1
	mu := mq.muR1
	notEmpty := mq.notEmptyR1[s]
	if r == 2 {
		msgs = mq.messagesR2
		mu = mq.muR2
		notEmpty = mq.notEmptyR2[s]
	}

	mu.Lock()
	defer mu.Unlock()

	msgQueue := msgs[s]
	msgQueue = append(msgQueue, msg)
	if msg.r == 1 {
		mq.messagesR1[msg.s] = msgQueue
	} else {
		mq.messagesR2[msg.s] = msgQueue
	}

	notEmpty.Broadcast()
}

func (mq *MessageQueue) Dequeue(r int, s int) *Message {
	msgs := mq.messagesR1
	notEmpty := mq.notEmptyR1[s]
	if r == 2 {
		msgs = mq.messagesR2
		notEmpty = mq.notEmptyR2[s]
	}

	notEmpty.L.Lock()
	defer notEmpty.L.Unlock()
	for len(msgs[s]) == 0 {
		notEmpty.Wait()
	}

	msg := msgs[s][0]
	msgs[s] = msgs[s][1:]
	if msg.r == 1 {
		mq.messagesR1[msg.s] = msgs[s]
	} else {
		mq.messagesR2[msg.s] = msgs[s]
	}

	return msg
}
