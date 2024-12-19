package main

import (
	"flag"
	"fmt"
	"github.com/schollz/progressbar/v3"
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

		log.Debug(Fields{
			"from": p.i,
			"to":   process.i,
			"data": msg,
		}, "Message sent")
	}
}

func (p *Process) gather() [MsgTypes]*atomic.Uint64 {
	log.Debug(Fields{
		"p": p.i,
		"r": p.r,
		"s": p.s,
	}, "Gathering")

	return p.msgQueue.DequeueEnoughMsg(p.r, p.s)
}

//goland:noinspection t
func benOr(p *Process, S int, f int, fCount *atomic.Uint64, odds float64, bar *progressbar.ProgressBar) {
	x := p.v
	var y V = NULL

	fUint64 := uint64(f)
	majority := uint64(len(*(p.processes))/2) + 1

	for p.s = 0; p.s < S; p.s++ {
		progressAdd(bar, 1) // progress bar

		if shouldStop(fUint64, fCount, odds) {
			p.stopped = true
			log.Debug(Fields{
				"p": p.i,
				"s": p.s,
			}, "STOPPED")
			return
		}

		// ###### Round 1 ######
		p.r = 1

		log.Debug(Fields{
			"p": p.i,
			"r": p.r,
			"s": p.s,
		}, "START PHASE")

		p.broadcast(x)
		countR1 := p.gather()

		for i := range MsgTypes {
			if countR1[i].Load() >= majority {
				if i == NullPos {
					y = V(NULL)
				} else {
					y = V(i)
				}
				break
			} else {
				y = NULL
			}
		}

		// ###### Round 2 ######
		p.r = 2

		log.Debug(Fields{
			"p": p.i,
			"r": p.r,
			"s": p.s,
		}, "START PHASE")

		p.broadcast(y)
		countR2 := p.gather()

		msgsAllNULL := true
		// range up to 1 because we do not want NULL values here, which is at index 2
		for i := range MsgTypes - 1 {
			if countR2[i].Load() >= fUint64+1 {
				p.decision = V(i)

				log.Debug(Fields{
					"p":        p.i,
					"decision": p.decision,
					"s":        p.s,
				}, "DECIDED")

				// you have to send the values to the next phase because some processes may need them to terminate
				// otherwise, a rare deadlock may happen
				p.r = 1
				p.s = p.s + 1
				p.broadcast(p.decision)
				p.r = 2
				p.broadcast(p.decision)
				p.s = p.s - 1

				return

			} else if countR2[i].Load() > 0 {
				x = V(i)
				msgsAllNULL = false
			}
		}

		// if all the messages were NULL
		if msgsAllNULL {
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
			messagesR1: make([][MsgTypes]*atomic.Uint64, S+1),
			messagesR2: make([][MsgTypes]*atomic.Uint64, S+1),

			enoughMsg:       uint64(n - f),
			enoughMsgCondR1: make([]*sync.Cond, S+1),
			enoughMsgCondR2: make([]*sync.Cond, S+1),
		}

		for s := range S + 1 {
			msgQueue.messagesR1[s] = [MsgTypes]*atomic.Uint64{{}, {}, {}}
			msgQueue.messagesR2[s] = [MsgTypes]*atomic.Uint64{{}, {}, {}}

			msgQueue.enoughMsgCondR1[s] = sync.NewCond(&sync.Mutex{})
			msgQueue.enoughMsgCondR2[s] = sync.NewCond(&sync.Mutex{})
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

var log = Logger{}

//goland:noinspection t
func main() {
	var n, f, S int

	flag.IntVar(&n, "n", 10, "number of processes")
	flag.IntVar(&f, "f", 4, "max number of stops")
	flag.IntVar(&S, "S", 10, "number of phases")
	odds := flag.Float64("odds", 0.05, "the odds of a process to stop. Valid values from 0.0 to 1.0")
	threads := flag.Int("threads", runtime.NumCPU(), "number of threads to use. Defaults to number of vCPU")
	csv := flag.Bool("csv", false, "print the the stats in csv format. Headers: n,f,fCount,S,maxS,decision,countViEQ0,countViEQ1")
	verbose := flag.Bool("verbose", false, "print all the messages sent and received in real time")
	disableProgressBar := flag.Bool("no-progress", false, "disable the progress bar")
	quite := flag.Bool("quite", false, "no output")
	initVals := flag.String("v", "", "initial values of the processes. Example: 1 0 1 1")
	flag.Parse()

	runtime.GOMAXPROCS(*threads)

	if *odds < 0.0 || *odds > 1.0 {
		log.Fatalf("Odds must be a value between 0.0 and 1.0. \"%f\" is not valid.", *odds)
	}

	if !*quite && *verbose {
		log.verbose = true
		log.Init()
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
		vi = randomVi(n)
	}

	if !(n > 2*f) {
		log.Fatalf("Error: n > 2f is not respected. n: %v, f: %v. Max f values must be: %v\n", n, f, int(n/2)-1)
	}

	var bar *progressbar.ProgressBar = nil
	if !*disableProgressBar && !*quite && !*csv {
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
			benOr((*processes)[i], S, f, fCount, *odds, bar)
		}()
	}

	wg.Wait()

	if *quite {
		return
	}

	if *csv {
		decision := V(NULL)
		maxPhases := 1
		countViEQ0 := 0
		countViEQ1 := 0
		for _, process := range *processes {
			if process.decision != NULL {
				decision = process.decision
			}
			if process.s > maxPhases {
				maxPhases = process.s
			}
			if process.v == 0 {
				countViEQ0 += 1
			} else {
				countViEQ1 += 1
			}
		}

		fmt.Printf("%d,%d,%d,%d,%d,%d,%d,%d\n", n, f, fCount.Load(), S, maxPhases, decision, countViEQ0, countViEQ1)

		return
	}

	fmt.Println("----- INIT VALUES -----")
	for i, v := range vi {
		fmt.Printf("v_%v: %v\n", i, v)
	}

	decided := false
	fmt.Println("----- DECISIONS -----")
	maxPhases := 1
	for _, process := range *processes {
		if process.stopped {
			fmt.Printf("P_%v stopped ", process.i)
		} else {
			if !decided && process.decision != NULL {
				decided = true
			}
			if process.s > maxPhases {
				maxPhases = process.s
			}
			fmt.Printf("P_%v decided: %v ", process.i, process.decision)
		}
		whenDecided := process.s
		if whenDecided == 0 {
			whenDecided = 1
		}
		fmt.Printf("at s:%d\n", whenDecided)
	}

	fmt.Println("----- INFO -----")
	terminateProbability := 1 - math.Pow(1-(1/math.Pow(2, float64(n))), float64(S))
	fmt.Printf("n: %d, f: %d, S: %d, majority: %d, termProb:%.2f%%, fCount: %v, odds of stopping: %f%%\n", n, f, S, int(n/2)+1, terminateProbability*100, fCount.Load(), *odds)

	if !decided {
		fmt.Print("Did NOT decide ")
	} else {
		fmt.Print("Decided ")
	}
	fmt.Printf("after %v/%v (%.2f%%) phases.\n", maxPhases, S, float64(maxPhases*100)/float64(S))
}
