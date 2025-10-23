package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ivansevryukov1995/golang_web_services_course/3/model"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	/*
		!!! !!! !!!
		обратите внимание - в задании обязательно нужен отчет
		делать его лучше в самом начале, когда вы видите уже узкие места, но еще не оптимизировалм их
		так же обратите внимание на команду в параметром -http
		перечитайте еще раз задание
		!!! !!! !!!
	*/

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)

	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := strings.Builder{}
	counter := -1

	user := &model.User{}

	for scanner.Scan() {
		counter++

		err := user.UnmarshalJSON(scanner.Bytes())
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for j := range user.Browsers {
			isAndroid = isAndroid || strings.Contains(user.Browsers[j], "Android")
			isMSIE = isMSIE || strings.Contains(user.Browsers[j], "MSIE")

			if strings.Contains(user.Browsers[j], "Android") || strings.Contains(user.Browsers[j], "MSIE") {
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == user.Browsers[j] {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers = append(seenBrowsers, user.Browsers[j])
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		email := strings.ReplaceAll(user.Email, "@", " [at] ")
		foundUsers.WriteString(fmt.Sprintf("[%d] %s <%s>\n", counter, user.Name, email))

	}

	fmt.Fprintln(out, "found users:\n"+foundUsers.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
