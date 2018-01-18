# Optimization of the program

The purpose of this assignment is to use pprof go-utility for identify problems in efficiency of the program and its optimization.
To accompllish this task is necesssary to achieve the following results: one of parameter should be better than BenchmarkSolution and another should be better than BenchmarkSolution + 20% ( fast < solution * 1.2).

BenchmarkSlow-8 10 142703250 ns/op 336887900 B/op 284175 allocs/op // initial

BenchmarkSolution-8 500 2782432 ns/op 559910 B/op 10422 allocs/op // aim

-----------------------------------------------------------------------------

BenchmarkFast-8   	     500	   2646164 ns/op	 1386229 B/op	    8226 allocs/op // my result
