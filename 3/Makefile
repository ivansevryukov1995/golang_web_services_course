.PHONY: out cpu mem

out:
	go test -bench . -benchmem -cpuprofile=cpu.out -memprofile=mem.out -memprofilerate=1
cpu: out
	go tool pprof -http=:8083 3.test.exe cpu.out
mem: out
	go tool pprof -http=:8083 3.test.exe mem.out