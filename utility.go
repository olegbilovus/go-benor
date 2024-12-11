package main

import (
	"github.com/schollz/progressbar/v3"
	"math/rand/v2"
	"sync/atomic"
)

func shouldStop(f uint64, fCount *atomic.Uint64) bool {
	fCountLocal := fCount.Load()
	if fCountLocal < f {
		if stop := rand.Uint32()%21 == 0; stop && fCount.CompareAndSwap(fCountLocal, fCountLocal+1) {
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
