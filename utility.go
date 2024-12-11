package main

import (
	"github.com/schollz/progressbar/v3"
	"github.com/sirupsen/logrus"
	"math/rand/v2"
	"sync/atomic"
)

func shouldStop(f uint64, fCount *atomic.Uint64, odds float64) bool {
	fCountLocal := fCount.Load()
	if fCountLocal < f {
		if stop := rand.Float64() <= odds; stop && fCount.CompareAndSwap(fCountLocal, fCountLocal+1) {
			return true
		}
	}

	return false
}

func progressAdd(bar *progressbar.ProgressBar, i int) {
	if bar != nil {
		_ = bar.Add(i)
	}
}

func resetCount(count map[V]int) {
	count[0] = 0
	count[1] = 0
	count[NULL] = 0
}

func randomVi(n int) []V {
	vi := make([]V, n)
	for i := 0; i < n; i++ {
		viRand := 0
		if rand.Int()%2 == 0 {
			viRand = 1
		}
		vi[i] = V(viRand)
	}

	return vi
}

type Logger struct {
	verbose bool
}

type Fields map[string]interface{}

func (l *Logger) Init() {
	if l.verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
func (l *Logger) Debug(fields map[string]interface{}, msg string) {
	if l.verbose {
		logrus.WithFields(fields).Debugln(msg)
	}
}

func (l *Logger) Fatalln(args ...interface{}) {
	logrus.Fatalln(args...)
}

func (l *Logger) Fatalf(s string, args ...interface{}) {
	logrus.Fatalf(s, args...)
}
