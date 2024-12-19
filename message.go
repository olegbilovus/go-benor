package main

import (
	"fmt"
	"sync"
	"sync/atomic"
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

const NULL_POS = 2

type MessageQueue struct {
	messagesR1 [][]*atomic.Uint64
	messagesR2 [][]*atomic.Uint64

	enoughMsg       uint64
	enoughMsgCondR1 []*sync.Cond
	enoughMsgCondR2 []*sync.Cond
}

func (mq *MessageQueue) HasEnoughMsgs(r int, s int) bool {
	msgs := mq.messagesR1
	if r == 2 {
		msgs = mq.messagesR2
	}

	if msgs[s][0].Load()+msgs[s][1].Load()+msgs[s][NULL_POS].Load() >= mq.enoughMsg {
		return true
	}

	return false
}

func (mq *MessageQueue) Enqueue(msg *Message) {
	r, s := msg.r, msg.s

	msgs := mq.messagesR1
	enoughMsgCond := mq.enoughMsgCondR1[s]
	if r == 2 {
		msgs = mq.messagesR2
		enoughMsgCond = mq.enoughMsgCondR2[s]
	}

	i := msg.v
	if msg.v == NULL {
		i = NULL_POS
	}
	msgs[s][i].Add(1)

	if mq.HasEnoughMsgs(r, s) {
		enoughMsgCond.Broadcast()
	}
}

func (mq *MessageQueue) DequeueEnoughMsg(r int, s int) []*atomic.Uint64 {
	msgs := mq.messagesR1
	enoughMsgCond := mq.enoughMsgCondR1[s]
	if r == 2 {
		msgs = mq.messagesR2
		enoughMsgCond = mq.enoughMsgCondR2[s]
	}

	enoughMsgCond.L.Lock()
	defer enoughMsgCond.L.Unlock()
	for !mq.HasEnoughMsgs(r, s) {
		enoughMsgCond.Wait()
	}

	return msgs[s]
}
