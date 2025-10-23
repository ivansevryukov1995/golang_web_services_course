.PHONY: out cpu mem

test:
	go test -v

out:
	go test -bench . -benchmem -cpuprofile=cpu.out -memprofile=mem.out -memprofilerate=1
out-after:
	go test -bench . -benchmem -cpuprofile=cpu_after.out -memprofile=mem_after.out -memprofilerate=1

cpu: out
	go tool pprof -http=:8083 3.test.exe cpu.out
mem: out
	go tool pprof -http=:8083 3.test.exe mem.out

cpu-comp:
	go tool pprof -http=:8083 -base cpu.out cpu_after.out
mem-comp:
	go tool pprof -http=:8083 -base mem.out mem_after.out