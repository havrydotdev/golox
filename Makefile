build:
	go build -o bin/golox main.go

profile:
	go test -bench='.' -count=10 -cpuprofile='cpu.prof' -memprofile='mem.prof'

test:
	go test -v ./...

fib_while: build
	./bin/golox ./_examples/fib_while.lox
