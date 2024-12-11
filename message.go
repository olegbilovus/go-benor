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

	enoughMsgR1 map[int]*sync.Cond
	enoughMsgR2 map[int]*sync.Cond
}

func (mq *MessageQueue) Enqueue(msg *Message) {
	r, s := msg.r, msg.s

	msgs := mq.messagesR1
	mu := mq.muR1
	enoughMsg := mq.enoughMsgR1[s]
	if r == 2 {
		msgs = mq.messagesR2
		mu = mq.muR2
		enoughMsg = mq.enoughMsgR2[s]
	}

	mu.Lock()
	defer mu.Unlock()

	msgQueue, ok := msgs[s]

	// it does not exist only if the msgQueue has already been dequeued when there were enough messages
	if !ok {
		return
	}

	msgQueue = append(msgQueue, msg)
	if msg.r == 1 {
		mq.messagesR1[msg.s] = msgQueue
	} else {
		mq.messagesR2[msg.s] = msgQueue
	}

	if len(msgQueue) >= n-f {
		enoughMsg.Broadcast()
	}
}

func (mq *MessageQueue) DequeueEnoughMsg(r int, s int) []*Message {
	msgs := mq.messagesR1
	enoughMsg := mq.enoughMsgR1[s]
	if r == 2 {
		msgs = mq.messagesR2
		enoughMsg = mq.enoughMsgR2[s]
	}

	enoughMsg.L.Lock()
	defer enoughMsg.L.Unlock()
	for len(msgs[s]) < n-f {
		enoughMsg.Wait()
	}

	msgsDequeued := msgs[s]
	delete(msgs, s)

	return msgsDequeued
}
