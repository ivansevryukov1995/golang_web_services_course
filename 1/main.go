package main

import (
	"fmt"
	"log"
	"os"
)

func dirTree(out *os.File, path string, printFiles bool) error {
	var prefix, sizeSufix string

	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}

	for i, file := range files {
		if i == len(files)-1 {
			prefix = "└───"
		} else {
			prefix = "├───"
		}

		if !file.IsDir() && file.Size() != 0 {
			sizeSufix = fmt.Sprintf("(%db)", file.Size())
		} else if !file.IsDir() && file.Size() == 0 {
			sizeSufix = "(empty)"
		} else {
			sizeSufix = ""
		}

		fmt.Printf("%s%s %s\n", prefix, file.Name(), sizeSufix)

		if file.IsDir() {
			// fmt.Printf("\t")
			dirTree(out, path+string(os.PathSeparator)+file.Name(), printFiles)
		}

	}

	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
