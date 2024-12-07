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
	messages []*Message
	mu       sync.Mutex
	notEmpty *sync.Cond
}

func (mq *MessageQueue) Add(msg *Message) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.messages = append(mq.messages, msg)
	mq.notEmpty.Broadcast()

}

func (mq *MessageQueue) Pop() *Message {
	mq.notEmpty.L.Lock()
	defer mq.mu.Unlock()

	for !(len(mq.messages) > 1) {
		mq.notEmpty.Wait()
	}

	msg := mq.messages[0]
	mq.messages = mq.messages[1:]

	return msg
}
