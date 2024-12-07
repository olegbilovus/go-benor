package main

import (
	"flag"
	"fmt"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
)

type V int8

const NULL = -1

//goland:noinspection t
func benOr(v V, p int) {
	x := v
	var y V = NULL
	for s := 1; s <= S; s++ {
		_ = bar.Add(1) // progress bar
		// ###### Round 1 ######
		log.Debugf("###### %v START r:%v s:%v", p, 1, s)
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
		log.Debugf("###### %v START r:%v s:%v", p, 2, s)
		broadcast(p, 2, s, y)
		msgsR2 := gather(p, 2, s)

		countR2 := map[V]int{}
		for _, msg := range msgsR2 {
			countR2[msg.v] += 1
			if countR2[msg.v] >= majority && msg.v != NULL {
				log.Debugf("P%v DECIDED: %v\n", p, msg)
				pDecisions[p] = msg.v
				break
			} else if msg.v != NULL {
				x = msg.v
			}
		}

		// if all the messages where NULL
		if countR2[NULL] == len(msgsR2) {
			x = V(0)
			if rand.Int()%2 == 0 {
				x = V(1)
			}

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
		log.Debugf("%v sent %v to %v\n", p, msg, i)
	}
}

func gather(p int, r int, s int) []*Message {
	var msgs []*Message

	msgQueue := pMessageQueues[p]

	for len(msgs) < n-f {
		msg := msgQueue.Pop()
		if msg.r == r && msg.s == s {
			msgs = append(msgs, msg)
			log.Debugf("%v received %v from %v\n", p, msg, msg.p)
		} else {
			log.Debugf("%v discarted %v from %v\n", p, msg, msg.p)
		}

	}

	return msgs
}

var n int
var f int
var S int
var majority int
var verbose bool

var pMessageQueues []*MessageQueue
var pDecisions []V

var bar *progressbar.ProgressBar

//goland:noinspection t
func main() {
	flag.IntVar(&n, "n", 3, "number of processors")
	flag.IntVar(&f, "f", 1, "max number of stops")
	flag.IntVar(&S, "S", 10, "number of phases")
	flag.BoolVar(&verbose, "verbose", false, "print all the messages sent and received in real time")
	initVals := flag.String("v", "", "initial values of the processors. Example: 1 0 1 1")
	flag.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
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
		n = len(sliceInitVals)
	} else {
		for i := 0; i < n; i++ {
			viRand := 0
			if rand.Int()%2 == 0 {
				viRand = 1
			}
			vi = append(vi, V(viRand))
		}
	}

	if !(n > 2*f) {
		log.Fatalf("Error: n > 2f is not respected. n: %v, f: %v. Max f values must be: %v\n", n, f, int(math.Floor(float64(n/2)))-1)
	}

	majority = int(math.Floor(float64(n/2)) + 1)

	// init global vars
	bar = progressbar.Default(int64(n * S))

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

	fmt.Println("----- INFO -----")
	fmt.Printf("n: %d, f: %d, majority: %d\n", n, f, majority)

	fmt.Println("----- INIT VALUES -----")
	for i, v := range vi {
		fmt.Printf("v_%v: %v\n", i, v)
	}

	fmt.Println("----- DECISIONS -----")
	for i, decision := range pDecisions {
		fmt.Printf("P_%v decided: %v\n", i, decision)
	}
}
