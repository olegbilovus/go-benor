package main

import (
	"flag"
	"fmt"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand/v2"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type V int

const NULL = -1

type Process struct {
	i        int
	v        V
	s        int
	r        int
	decision V
	stopped  bool
	msgQueue *MessageQueue

	processes *[]*Process
}

func (p *Process) broadcast(v V) {
	msg := &Message{
		r: p.r,
		s: p.s,
		v: v,
		p: p.i,
	}

	for _, process := range *p.processes {
		if process.stopped || process.s > p.s {
			continue
		}
		process.msgQueue.Enqueue(msg)

		log.WithFields(log.Fields{
			"from": p.i,
			"to":   process.i,
			"data": msg,
		}).Debugln("Message sent")
	}
}

func (p *Process) gather() []*Message {
	msgQueue := p.msgQueue
	msgs := msgQueue.DequeueEnoughMsg(p.r, p.s)

	for _, msg := range msgs {
		log.WithFields(log.Fields{
			"from": msg.p,
			"to":   p.i,
			"data": msg,
		}).Debugln("Message received")

	}

	return msgs
}

//goland:noinspection t
func benOr(p *Process, S int, f int, fCount *atomic.Uint64, bar *progressbar.ProgressBar) {
	x := p.v
	var y V = NULL

	fUint64 := uint64(f)
	majority := int(len(*(p.processes))/2) + 1

	count := make(map[V]int, 3) // 3 because it will only be 0, 1, -1

	for p.s = 0; p.s < S; p.s++ {
		progressAdd(bar, 1) // progress bar

		if shouldStop(fUint64, fCount) {
			p.stopped = true
			progressAdd(bar, S-p.s)
			return
		}

		// ###### Round 1 ######
		p.r = 1

		log.WithFields(log.Fields{
			"p": p.i,
			"r": p.r,
			"s": p.s,
		}).Debugln("START PHASE")

		p.broadcast(x)
		msgsR1 := p.gather()

		resetCount(count)
		for _, msg := range msgsR1 {
			count[msg.v] += 1
			if count[msg.v] >= majority {
				y = msg.v
				break
			} else {
				y = NULL
			}
		}

		// ###### Round 2 ######
		p.r = 2

		log.WithFields(log.Fields{
			"p": p.i,
			"r": p.r,
			"s": p.s,
		}).Debugln("START PHASE")

		p.broadcast(y)
		msgsR2 := p.gather()

		resetCount(count)
		for _, msg := range msgsR2 {
			count[msg.v] += 1
			if msg.v != NULL && count[msg.v] >= f+1 {
				p.decision = msg.v

				log.WithFields(log.Fields{
					"p":        p.i,
					"decision": p.decision,
					"s":        p.s,
				}).Debugln("DECIDED")

				return

			} else if msg.v != NULL {
				x = msg.v
			}
		}

		// if all the messages where NULL
		if count[NULL] == len(msgsR2) {
			x = V(0)
			if rand.Int()%2 == 0 {
				x = V(1)
			}

		}

	}
}

func SetupProcesses(n int, f int, S int, vi []V) *[]*Process {
	processes := make([]*Process, n)
	for i := 0; i < n; i++ {
		msgQueue := &MessageQueue{
			messagesR1: make(map[int][]*Message, S),
			messagesR2: make(map[int][]*Message, S),

			muR1: &sync.Mutex{},
			muR2: &sync.Mutex{},

			enoughMsg:       n - f,
			enoughMsgCondR1: make(map[int]*sync.Cond, S),
			enoughMsgCondR2: make(map[int]*sync.Cond, S),
		}

		for s := range S {
			msgQueue.messagesR1[s] = make([]*Message, 0)
			msgQueue.messagesR2[s] = make([]*Message, 0)
			msgQueue.enoughMsgCondR1[s] = sync.NewCond(msgQueue.muR1)
			msgQueue.enoughMsgCondR2[s] = sync.NewCond(msgQueue.muR2)
		}

		process := &Process{
			i:         i,
			v:         vi[i],
			s:         0,
			r:         1,
			decision:  NULL,
			stopped:   false,
			msgQueue:  msgQueue,
			processes: &processes,
		}
		processes[i] = process
	}

	return &processes
}

//goland:noinspection t
func main() {
	var n, f, S int

	flag.IntVar(&n, "n", 3, "number of processes")
	flag.IntVar(&f, "f", 1, "max number of stops")
	flag.IntVar(&S, "S", 10, "number of phases")
	threads := flag.Int("threads", runtime.NumCPU(), "number of threads to use. Defaults to number of vCPU")
	verbose := flag.Bool("verbose", false, "print all the messages sent and received in real time")
	disableProgressBar := flag.Bool("no-progress", false, "disable the progress bar")
	initVals := flag.String("v", "", "initial values of the processes. Example: 1 0 1 1")
	flag.Parse()

	runtime.GOMAXPROCS(*threads)

	if *verbose {
		log.SetLevel(log.DebugLevel)
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
		log.Fatalf("Error: n > 2f is not respected. n: %v, f: %v. Max f values must be: %v\n", n, f, int(n/2)-1)
	}

	var bar *progressbar.ProgressBar = nil
	if !*disableProgressBar {
		bar = progressbar.Default(int64(n * S))
	}

	// init processes
	processes := SetupProcesses(n, f, S, vi)

	fCount := &atomic.Uint64{}
	// start computation
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			benOr((*processes)[i], S, f, fCount, bar)
		}()
	}

	wg.Wait()

	fmt.Println("----- INIT VALUES -----")
	for i, v := range vi {
		fmt.Printf("v_%v: %v\n", i, v)
	}

	decided := false
	fmt.Println("----- DECISIONS -----")
	maxStates := 0
	for _, process := range *processes {
		if process.stopped {
			fmt.Printf("P_%v stopped\n", process.i)
		} else {
			if !decided && process.decision != NULL {
				decided = true
			}
			if process.s > maxStates {
				maxStates = process.s
			}
			fmt.Printf("P_%v decided: %v\n", process.i, process.decision)
		}
	}

	fmt.Println("----- INFO -----")
	terminateProbability := 1 - math.Pow(1-(1/math.Pow(2, float64(n))), float64(S))
	fmt.Printf("n: %d, f: %d, S: %d, majority: %d, termProb:%.2f%%, fCount: %v\n", n, f, S, int(n/2)+1, terminateProbability*100, fCount.Load())

	if !decided {
		fmt.Print("Did NOT decide ")
	} else {
		fmt.Print("Decided ")
	}
	fmt.Printf("after %v/%v (%.2f%%) phases.\n", maxStates, S, float64(maxStates*100)/float64(S))
}
