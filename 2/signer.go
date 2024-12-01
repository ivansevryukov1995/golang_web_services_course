package main

func SingleHash(data string) string {
	s := DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
	return s
}

func MultiHash() {

}

// func ExecutePipeline(freeFlowJobs []job) {

// }

func main() {
	data := "0"
	SingleHash(data)
}
