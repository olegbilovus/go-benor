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
	"sync/atomic"
)

type V int8

const NULL = -1

type Process struct {
	i        int
	v        V
	s        int
	r        int
	decision V
	stopped  bool
	msgQueue *MessageQueue

	processes []*Process
}

func (p *Process) broadcast(v V) {
	msg := &Message{
		r: p.r,
		s: p.s,
		v: v,
		p: p.i,
	}

	for _, process := range p.processes {
		process.msgQueue.Enqueue(msg)

		log.WithFields(log.Fields{
			"from": p.i,
			"to":   process.i,
			"data": msg,
		}).Debugln("Message sent")
	}
}

func (p *Process) gather() []*Message {
	var msgs []*Message
	msgQueue := p.msgQueue

	for len(msgs) < n-f {
		msg := msgQueue.Dequeue(p.r, p.s)
		msgs = append(msgs, msg)

		log.WithFields(log.Fields{
			"from": msg.p,
			"to":   p.i,
			"data": msg,
		}).Debugln("Message received")

	}

	return msgs
}

func shouldStop() bool {
	fCountLocal := fCount.Load()
	if fCountLocal < fUint64 {
		if stop := rand.Uint32()%21 == 0; stop && fCount.CompareAndSwap(fCountLocal, fCountLocal+1) {
			return true
		}
	}

	return false
}

//goland:noinspection t
func benOr(p *Process) {
	x := p.v
	var y V = NULL
	for p.s = 0; p.s < S; p.s++ {
		_ = bar.Add(1) // progress bar

		if shouldStop() {
			p.stopped = true
			_ = bar.Add(S - p.s)
			return
		}

		// ###### Round 1 ######
		p.r = 1
		log.Debugf("###### %v START r:%v s:%v", p.i, p.r, p.s)
		p.broadcast(x)
		msgsR1 := p.gather()

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
		p.r = 2
		log.Debugf("###### %v START r:%v s:%v", p.i, p.r, p.s)
		p.broadcast(y)
		msgsR2 := p.gather()

		countR2 := map[V]int{}
		for _, msg := range msgsR2 {
			countR2[msg.v] += 1
			if msg.v != NULL && countR2[msg.v] >= f+1 {
				log.Debugf("P%v DECIDED: %v", p, msg)
				p.decision = msg.v
				x = msg.v

				return
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

var n int
var f int
var fUint64 uint64
var S int
var majority int
var verbose bool

var fCount atomic.Uint64

var bar *progressbar.ProgressBar

//goland:noinspection t
func main() {
	flag.IntVar(&n, "n", 3, "number of processes")
	flag.IntVar(&f, "f", 1, "max number of stops")
	flag.IntVar(&S, "S", 10, "number of phases")
	flag.BoolVar(&verbose, "verbose", false, "print all the messages sent and received in real time")
	initVals := flag.String("v", "", "initial values of the processes. Example: 1 0 1 1")
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
	fUint64 = uint64(f)

	majority = int(math.Floor(float64(n/2)) + 1)
	fCount.Store(0)

	bar = progressbar.Default(int64(n * S))

	// init processes
	processes := make([]*Process, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		msgQueue := &MessageQueue{
			messagesR1: make(map[int][]*Message, S),
			messagesR2: make(map[int][]*Message, S),

			muR1: &sync.Mutex{},
			muR2: &sync.Mutex{},

			notEmptyR1: make(map[int]*sync.Cond, S),
			notEmptyR2: make(map[int]*sync.Cond, S),
		}

		for s := range S {
			msgQueue.messagesR1[s] = make([]*Message, 0)
			msgQueue.messagesR2[s] = make([]*Message, 0)
			msgQueue.notEmptyR1[s] = sync.NewCond(msgQueue.muR1)
			msgQueue.notEmptyR2[s] = sync.NewCond(msgQueue.muR2)
		}

		process := &Process{
			i:         i,
			v:         vi[i],
			s:         0,
			r:         1,
			decision:  NULL,
			stopped:   false,
			msgQueue:  msgQueue,
			processes: processes,
		}
		processes[i] = process
	}

	// start computation
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			benOr(processes[i])
		}()
	}

	wg.Wait()

	fmt.Println("----- INIT VALUES -----")
	for i, v := range vi {
		fmt.Printf("v_%v: %v\n", i, v)
	}

	decided := false
	fmt.Println("----- DECISIONS -----")
	for _, process := range processes {
		if process.stopped {
			fmt.Printf("P_%v stopped\n", process.i)
		} else {
			decided = true
			fmt.Printf("P_%v decided: %v\n", process.i, process.decision)
		}
	}

	fmt.Println("----- INFO -----")
	terminateProbability := 1 - math.Pow(1-(1/math.Pow(2, float64(n))), float64(S))
	fmt.Printf("n: %d, f: %d, S: %d, majority: %d, termProb:%.2f%%, fCount: %v\n", n, f, S, majority, terminateProbability*100, fCount.Load())

	if !decided {
		fmt.Print("Did NOT decided ")
	} else {
		fmt.Print("Decided ")
	}
	fmt.Printf("after %v/%v (%.2f%%) phases.\n", bar.State().CurrentNum/int64(n), S, bar.State().CurrentPercent*100)
}
