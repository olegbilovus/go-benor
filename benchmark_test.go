package main

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
	"testing"
)

func benOrBench(S int, n int) {
	f := n/2 - 1
	fCount := &atomic.Uint64{}
	if !(n > 2*f) {
		log.Fatalln("Error property: n > 2f")
	}

	processes := SetupProcesses(n, f, S, randomVi(n))
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			benOr((*processes)[i], S, f, fCount, nil)
		}()
	}

	wg.Wait()
}

func BenchmarkBenOr_N10S5(b *testing.B) {
	log.SetLevel(log.InfoLevel)
	for range b.N {
		benOrBench(5, 10)
	}
}

func BenchmarkBenOr_N10S50(b *testing.B) {
	log.SetLevel(log.InfoLevel)
	for range b.N {
		benOrBench(50, 10)
	}
}

func BenchmarkBenOr_N10S500(b *testing.B) {
	log.SetLevel(log.InfoLevel)
	for range b.N {
		benOrBench(500, 10)
	}
}

func BenchmarkBenOr_N50S5(b *testing.B) {
	log.SetLevel(log.InfoLevel)
	for range b.N {
		benOrBench(5, 50)
	}
}

func BenchmarkBenOr_N50S50(b *testing.B) {
	log.SetLevel(log.InfoLevel)
	for range b.N {
		benOrBench(50, 50)
	}
}

func BenchmarkBenOr_N50S500(b *testing.B) {
	log.SetLevel(log.InfoLevel)
	for range b.N {
		benOrBench(500, 50)
	}
}

func BenchmarkBenOr_N500S5(b *testing.B) {
	log.SetLevel(log.InfoLevel)
	for range b.N {
		benOrBench(5, 500)
	}
}

func BenchmarkBenOr_N500S50(b *testing.B) {
	log.SetLevel(log.InfoLevel)
	for range b.N {
		benOrBench(50, 500)
	}
}

func BenchmarkBenOr_N500S500(b *testing.B) {
	log.SetLevel(log.InfoLevel)
	for range b.N {
		benOrBench(500, 500)
	}
}
