package main

import (
	"fmt"
	"sync"
)

type Message struct {
	r        int
	s        int
	v        V
	sender   int
	receiver int
}

func (m *Message) String() string {
	return fmt.Sprintf("(r:%v, s:%v, v:%v)", m.r, m.s, m.v)
}

const NullPos = 2
const MsgTypes = 3

type MessageQueue struct {
	messagesR1 [][MsgTypes]uint64
	messagesR2 [][MsgTypes]uint64

	enoughMsg       uint64
	enoughMsgCondR1 []*sync.Cond
	enoughMsgCondR2 []*sync.Cond
}

func (mq *MessageQueue) HasEnoughMsgs(r int, s int) bool {
	msgs := mq.messagesR1
	if r == 2 {
		msgs = mq.messagesR2
	}

	return msgs[s][0]+msgs[s][1]+msgs[s][NullPos] >= mq.enoughMsg
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
		i = NullPos
	}

	enoughMsgCond.L.Lock()
	msgs[s][i] += 1
	enoughMsgCond.L.Unlock()

	if mq.HasEnoughMsgs(r, s) {
		enoughMsgCond.Broadcast()
	}
}

func (mq *MessageQueue) DequeueEnoughMsg(r int, s int) [MsgTypes]uint64 {
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
