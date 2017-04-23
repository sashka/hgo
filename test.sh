#!/bin/bash

go test -run none . -bench . -benchtime 3s  -benchmem -memprofile mem.pprof -cpuprofile cpu.pprof

echo "----> CPU Profile:"
go tool pprof -text -call_tree ./hgo.test cpu.pprof