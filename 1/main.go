package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func dirTree(out io.Writer, path string, printFiles bool) error {
	var prefix, sizeSufix, endFile string

	dir, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer dir.Close()

	files, err := dir.Readdir(0)
	if err != nil {
		log.Fatal(err)
	}

	if !printFiles {
		tempFiles := []os.FileInfo{}
		for _, file := range files {
			if file.IsDir() {
				tempFiles = append(tempFiles, file)
			}
		}
		files = tempFiles
	}

	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

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

		for j := 0; j < len(strings.Split(path+string(os.PathSeparator)+file.Name(), "\\"))-2; j++ {

			parent := filepath.Join(strings.Split(path+string(os.PathSeparator)+file.Name(), "\\")[:j+1]...)

			dirParent, err := os.Open(parent)
			if err != nil {
				log.Fatal(err)
			}
			defer dirParent.Close()

			filesParent, err := dirParent.Readdir(-1)
			if err != nil {
				log.Fatal(err)
			}

			if !printFiles {
				tempFiles := []os.FileInfo{}
				for _, fileParent := range filesParent {
					if fileParent.IsDir() {
						tempFiles = append(tempFiles, fileParent)
					}
				}
				filesParent = tempFiles
			}

			sort.Slice(filesParent, func(n, m int) bool { return filesParent[n].Name() < filesParent[m].Name() })

			for n, fileParent := range filesParent {
				if n == len(filesParent)-1 {
					endFile = fileParent.Name()
				}
			}

			if strings.Split(path+string(os.PathSeparator)+file.Name(), "\\")[j+1] != endFile {
				fmt.Printf("│\t")
			} else {
				fmt.Printf("\t")
			}

		}

		fmt.Printf("%s%s %s\n", prefix, file.Name(), sizeSufix)

		if file.IsDir() {
			err := dirTree(out, path+string(os.PathSeparator)+file.Name(), printFiles)
			if err != nil {
				log.Fatal(err)
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
