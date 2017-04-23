package main

import (
	"testing"

	"github.com/sashka/hgo/command"
)

func BenchmarkStatus100(b *testing.B) {
	status := command.StatusCommand{}
	for n := 0; n < b.N; n++ {
		status.Run(make([]string, 0))
	}
}
