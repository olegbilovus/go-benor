package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
)

type V int8
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
	mq.messages = append(mq.messages, msg)
	mq.notEmpty.Broadcast()
	mq.mu.Unlock()
}

func (mq *MessageQueue) Pop() *Message {
	mq.notEmpty.L.Lock()
	for !(len(mq.messages) > 1) {
		mq.notEmpty.Wait()
	}

	msg := mq.messages[0]
	mq.messages = mq.messages[1:]
	mq.mu.Unlock()

	return msg
}

const NULL = -1

//goland:noinspection t
func benOr(v V, p int) {
	x := v
	var y V = NULL

	for s := 1; s <= S; s++ {
		// ###### Round 1 ######
		log.Printf("###### %v START r:%v s:%v", p, 1, s)
		broadcast(p, 1, s, x)
		msgsR1 := gather(p, 1, s)

		countR1 := map[V]int{}
		for _, msg := range msgsR1 {
			countR1[msg.v] += 1
			if countR1[msg.v] >= majority {
				y = msg.v
				break
			} else {
				y = NULL
			}
		}

		// ###### Round 2 ######
		log.Printf("###### %v START r:%v s:%v", p, 2, s)
		broadcast(p, 2, s, y)
		msgsR2 := gather(p, 2, s)

		countR2 := map[V]int{}
		for _, msg := range msgsR2 {
			countR2[msg.v] += 1
			if countR2[msg.v] >= majority && msg.v != NULL {
				log.Printf("P%v DECIDED: %v\n", p, msg)
				pDecisions[p] = msg.v
				break
			} else if msg.v != NULL {
				x = msg.v
			}
		}

		// if all the messages where NULL
		if countR2[NULL] == len(msgsR2) {
			x = V(rand.IntN(1))
		}

	}
}

func broadcast(p int, r int, s int, v V) {
	for i, pMsgQueue := range pMessageQueues {
		msg := &Message{
			r: r,
			s: s,
			v: v,
			p: p,
		}
		pMsgQueue.Add(msg)
		log.Printf("%v sent %v to %v\n", p, msg, i)
	}
}

func gather(p int, r int, s int) []*Message {
	var msgs []*Message

	msgQueue := pMessageQueues[p]

	for len(msgs) < n-f {
		msg := msgQueue.Pop()
		if msg.r == r && msg.s == s {
			msgs = append(msgs, msg)
			log.Printf("%v received %v from %v\n", p, msg, msg.p)
		} else {
			log.Printf("%v discarted %v from %v\n", p, msg, msg.p)
		}

	}

	return msgs
}

var n int
var f int
var S int
var majority int

var pMessageQueues []*MessageQueue
var pDecisions []V

//goland:noinspection t
func main() {
	flag.IntVar(&n, "n", 3, "number of processors")
	flag.IntVar(&f, "f", 1, "max number of stops")
	flag.IntVar(&S, "S", 10, "number of phases")
	initVals := flag.String("v", "", "initial values of the processors. Example: 1 0 1 1")
	flag.Parse()

	if !(n > 2*f) {
		log.Fatalln("Error: n > 2f is not respected")
	}

	var vi []V
	if *initVals != "" {
		sliceInitVals := strings.Split(*initVals, " ")
		for _, v := range sliceInitVals {
			vInt, err := strconv.Atoi(v)
			if err != nil || !(vInt == 0 || vInt == 1) {
				log.Fatalf("Only 0 and 1 are valid values for v. \"%v\" is not valid.", v)
			}
			vi = append(vi, V(vInt))
		}
	} else {
		for i := 0; i < n; i++ {
			viRand := 0
			if rand.Int()%2 == 0 {
				viRand = 1
			}
			vi = append(vi, V(viRand))
		}
	}

	majority = int(math.Floor(float64(n/2)) + 1)

	// init global vars
	pMessageQueues = make([]*MessageQueue, n)
	pDecisions = make([]V, n)
	for i := 0; i < len(pMessageQueues); i++ {
		msgQueue := &MessageQueue{
			messages: make([]*Message, 0),
			mu:       sync.Mutex{},
		}
		msgQueue.notEmpty = sync.NewCond(&msgQueue.mu)
		pMessageQueues[i] = msgQueue

		pDecisions[i] = V(-1)
	}

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			benOr(vi[i], i)
		}()
	}

	wg.Wait()

	log.Println("----- INFO -----")
	log.Printf("n: %d, f: %d, majority: %d\n", n, f, majority)

	log.Println("----- INIT VALUES -----")
	for i, v := range vi {
		log.Printf("v_%v: %v\n", i, v)
	}

	log.Println("----- DECISIONS -----")
	for i, decision := range pDecisions {
		log.Printf("%v decided: %v\n", i, decision)
	}
}
