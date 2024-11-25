package main

import (
	"fmt"
	"log"
	"os"
)

func dirTree(out *os.File, path string, printFiles bool) error {

	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	f, err := dir.Readdir(-1)
	if err != nil {
		log.Fatal(err)
	}

	for i, file := range f {
		if i == len(f)-1 {
			fmt.Printf("└───%s\n", file.Name())
			if file.IsDir() {
				fmt.Printf("\t")
				return dirTree(out, path+string(os.PathSeparator)+file.Name(), printFiles)
			}
		} else {
			fmt.Printf("├───%s\n", file.Name())
			if file.IsDir() {
				fmt.Printf("\t")
				return dirTree(out, path+string(os.PathSeparator)+file.Name(), printFiles)
			}
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
