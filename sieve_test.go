package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSieve(t *testing.T) {
	primes := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97}
	p := primeGenerator{100}
	output := p.Start()

	for _, p := range primes {
		q, open := <-output
		assert.True(t, open)
		assert.Equal(t, p, q)
	}
	_, open := <-output
	assert.False(t, open)
}
