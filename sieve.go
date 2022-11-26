package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) <= 1 {
		log.Fatal("must provide N as first arg")
	}

	strN := os.Args[1]
	N, err := strconv.Atoi(strN)
	if err != nil {
		log.Fatalf("arg %q is not a number", strN)
	}

	p := primeGenerator{N}
	primes := p.Start()
	for p := range primes {
		fmt.Println(p)
		time.Sleep(100 * time.Millisecond)
	}
}

// Prime generator
type primeGenerator struct {
	N int // upper limit
}

func (p *primeGenerator) Start() chan int {
	// primes will be sent on this channel
	primeCh := make(chan int, p.N)
	go p.getPrimes(primeCh)
	return primeCh
}

func (p *primeGenerator) getPrimes(primeCh chan int) {
	// sieve[p] == true  <=>  p is composite
	sieve := make([]bool, p.N+1)
	sieve[0] = true
	sieve[1] = true

	doneCh := make(chan int, sqrt(p.N))
	go p.filter(sieve, doneCh)

	// Start release worker
	releaseCh := make(chan int)
	releaserDoneCh := make(chan int)
	rl := releaser{p.N, sieve, releaseCh, primeCh, releaserDoneCh}
	go rl.listen()

	// Keep track of which filterers are done
	done := make([]bool, sqrt(p.N)+1)
	waitingFor := 2

	for waitingFor <= sqrt(p.N) {
		if done[waitingFor] {
			// All filters up to m := `waitingFor` are done, so we know all numbers
			// below m^2 in the sieve are prime. Tell the releaser to release them.
			releaseCh <- waitingFor
			waitingFor++
			continue
		}

		// Wait for next worker to finish
		nextDone := <-doneCh
		done[nextDone] = true
	}

	// Wait for releaser to finish
	<-releaserDoneCh

	close(primeCh)
}

func (p *primeGenerator) filter(sieve []bool, doneCh chan int) {
	for k := 2; k <= sqrt(p.N); k++ {
		// filterer k is responsible for filtering k*k, k*(k+1), ...
		f := filterer{sieve, k, p.N, doneCh}
		go f.filter()
	}
}

// Filterer workers - each filtering part of a sieve
type filterer struct {
	sieve  []bool
	k, N   int
	doneCh chan int
}

func (f *filterer) filter() {
	for i := f.k; f.k*i <= f.N; i++ {
		f.sieve[f.k*i] = true
	}
	f.doneCh <- f.k
}

// Release worker - waits for communication on a channel and then releases
// part of the sieve
type releaser struct {
	N     int
	sieve []bool

	inputCh, primeCh, doneCh chan int
}

func (r *releaser) listen() {
	for {
		select {
		case M := <-r.inputCh:
			r.release(M)
		case <-r.doneCh:
			return
		}
	}
}

// Send all primes in the sieve between (M-1)^2 and M^2 to the channel.
func (r *releaser) release(M int) {
	for q := (M - 1) * (M - 1); q <= M*M && q <= r.N; q++ {
		if !r.sieve[q] {
			r.primeCh <- q
		}
	}

	if M == sqrt(r.N) {
		// We're done
		r.doneCh <- 0
		r.doneCh <- 0
	}
}

// Helper functions
func sqrt(n int) int {
	return int(math.Sqrt(float64(n)))
}
