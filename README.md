<a id='anchor'></a>
# Golang web services course
Этот репозиторий является списком выполненных заданий [курса](https://stepik.org/course/187490/syllabus) с сайта Stepik
за авторством [Василия Романова](https://github.com/rvasily).

<img src=1\testdata\project\gopher.png height="200" width="200">

1. [Программа вывода дерева файлов](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/1)
<details><summary>Result:</summary>

```bash
=== RUN   TestTreeFull
--- PASS: TestTreeFull (0.00s)
=== RUN   TestTreeDir
--- PASS: TestTreeDir (0.00s)
PASS
ok      hw      0.157s
```
</details>

2. [Асинхронный пайплайн](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/2)
<details><summary>Result:</summary>

```bash
=== RUN   TestByIlia
collected 3
collected 9
collected 12
--- PASS: TestByIlia (0.30s)
=== RUN   TestPipeline
--- PASS: TestPipeline (0.01s)
=== RUN   TestSigner
--- PASS: TestSigner (2.07s)
PASS
ok      2       2.565s
```
</details>

3. [Оптимизация кода](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/3)
<details>
<summary>Result:</summary>

<details>
<summary>Замена regexp.MatchString на strings.Contains:</summary>

```go
// if ok, err := regexp.MatchString("Android", browser); ok && err == nil 
if strings.Contains(browser, "Android") 
и
// if ok, err := regexp.MatchString("MSIE", browser); ok && err == nil 
if strings.Contains(browser, "MSIE") 
```
Статистика

```
BenchmarkSlow-12               3         475009367 ns/op        20406149 B/op     182837 allocs/op
BenchmarkFast-12               8         136187475 ns/op         6216179 B/op      46750 allocs/op
```
</details>

<details>
<summary>Убрал map users. Все значения одновременно не нужны:</summary>

```go
// users := make([]map[string]interface{}, 0)

// for _, line := range lines {
// 	user := make(map[string]interface{})
// 	// fmt.Printf("%v %v\n", err, line)
// 	err := json.Unmarshal([]byte(line), &user)
// 	if err != nil {
// 		panic(err)
// 	}
// 	users = append(users, user)
// }

// for i, user := range users {
user := make(map[string]interface{})
for i, line := range lines {

    err := json.Unmarshal([]byte(line), &user)
    if err != nil {
        panic(err)
    }
    ...
```
Статистика

```
BenchmarkSlow-12               3         469516967 ns/op        20406941 B/op     182831 allocs/op
BenchmarkFast-12               9         127068733 ns/op         5892163 B/op      43753 allocs/op
```

</details>

<details>
<summary>Убрал преобразования типов, которые можно избежать, добавив struct User:</summary>

```go
type User struct {
	Browsers []string `json:"browsers"`
	Company  string   `json:"company"`
	Country  string   `json:"country"`
	Email    string   `json:"email"`
	Job      string   `json:"job"`
	Name     string   `json:"name"`
	Phone    string   `json:"phone"`
}
...
// user := make(map[string]interface{})
user := &User{}
...
// browsers, ok := user["browsers"].([]interface{})
browsers := user.Browsers
// if !ok {
// 	// log.Println("cant cast browsers")
// 	continue
// }

for _, browser := range browsers {
// browser, ok := browserRaw.(string)

// if !ok {
// 	// log.Println("cant cast browser to string")
// 	continue
// }
...
for _, browser := range browsers {
// browser, ok := browserRaw.(string)
// if !ok {
// 	// log.Println("cant cast browser to string")
// 	continue
// }
...
// email := r.ReplaceAllString(user["email"].(string), " [at] ")
// foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
email := r.ReplaceAllString(user.Email, " [at] ")
foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email)
...
    
```
Статистика

```
BenchmarkSlow-12               3         464770200 ns/op        20357064 B/op     182835 allocs/op
BenchmarkFast-12              22          50302523 ns/op         5476525 B/op      16835 allocs/op
```
</details>

<details>
<summary>Убрал дублирующий range по browsers. Производительность не увеличилась:</summary>

```go
// browsers := user.Browsers

// for _, browser := range browsers {
// 	if strings.Contains(browser, "Android") {
// 		isAndroid = true
// 		notSeenBefore := true
// 		for _, item := range seenBrowsers {
// 			if item == browser {
// 				notSeenBefore = false
// 			}
// 		}
// 		if notSeenBefore {
// 			// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
// 			seenBrowsers = append(seenBrowsers, browser)
// 			uniqueBrowsers++
// 		}
// 	}
// }

// for _, browser := range browsers {
// 	if strings.Contains(browser, "MSIE") {
// 		isMSIE = true
// 		notSeenBefore := true
// 		for _, item := range seenBrowsers {
// 			if item == browser {
// 				notSeenBefore = false
// 			}
// 		}
// 		if notSeenBefore {
// 			// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
// 			seenBrowsers = append(seenBrowsers, browser)
// 			uniqueBrowsers++
// 		}
// 	}
// }

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
...
```
Статистика

```
BenchmarkSlow-12               2         522537750 ns/op        20333724 B/op     182836 allocs/op
BenchmarkFast-12              22          52342009 ns/op         5479009 B/op      16835 allocs/op
```

</details>

<details>
<summary>Заменил чтение всего файла io.ReadAll чтением построчно bufio.NewScanner:</summary>

```go
// fileContents, err := io.ReadAll(file)
	// if err != nil {
	// 	panic(err)
	// }
// lines := strings.Split(string(fileContents), "\n")
for scanner.Scan() {
    err := json.Unmarshal([]byte(scanner.Text()), &user)
...
}
```
Статистика

```
BenchmarkSlow-12               3         476512567 ns/op        20243784 B/op     182810 allocs/op
BenchmarkFast-12              21          52663105 ns/op         2233568 B/op      17803 allocs/op
```
</details>

<details>
<summary>Заменил []byte(scanner.Text()) на scanner.Bytes():</summary>

```go
///err := json.Unmarshal([]byte(scanner.Text()), &user)
err := json.Unmarshal(scanner.Bytes(), &user)

```
Статистика

```
BenchmarkSlow-12               3         478741333 ns/op        20369272 B/op     182826 allocs/op
BenchmarkFast-12              24          49166408 ns/op         1015203 B/op      15793 allocs/op
```

</details>

<details>
<summary>Заменил r.ReplaceAllString на strings.ReplaceAll:</summary>

```go
// r := regexp.MustCompile("@")
// email := r.ReplaceAllString(user.Email, " [at] ")
email := strings.ReplaceAll(user.Email, "@", " [at] ")

```
Статистика

```
BenchmarkSlow-12               3         475121533 ns/op        20346645 B/op     182824 allocs/op
BenchmarkFast-12              24          48949983 ns/op          970686 B/op      15493 allocs/op
```

</details>

<details>
<summary>Заменил json.Unmarshal на сгенерированный метод UnmarshalJSON кодогенерацией из easyjson:</summary>

```go
// err := json.Unmarshal(, &user)
err := user.UnmarshalJSON(scanner.Bytes())
```
Статистика

```
BenchmarkSlow-12               2         534617250 ns/op        20447088 B/op     182847 allocs/op
BenchmarkFast-12              55          26175882 ns/op          735485 B/op      10486 allocs/op
```

</details>

<details>
<summary>Выставил флаг json:"-" для ненужных полей структуры User:</summary>

```go
type User struct {
	Browsers []string `json:"browsers"`
	Company  string   `json:"-"`
	Country  string   `json:"-"`
	Email    string   `json:"email"`
	Job      string   `json:"-"`
	Name     string   `json:"name"`
	Phone    string   `json:"-"`
}
```
Статистика

```
BenchmarkSlow-12              60          21951848 ns/op        20437123 B/op     182855 allocs/op
BenchmarkFast-12             736           1543819 ns/op          676168 B/op       6482 allocs/op
```

</details>
<details>
<summary>Заменил конкатенацию строк на strings.Builder:</summary>

```go
//foundUsers += fmt.Sprintf("[%d] %s <%s>\n", counter, user.Name, email)
foundUsers.WriteString(fmt.Sprintf("[%d] %s <%s>\n", counter, user.Name, email))
```
Статистика

```
BenchmarkSlow-12              54          24577407 ns/op        20480195 B/op     182861 allocs/op
BenchmarkFast-12             628           2131925 ns/op          505364 B/op       6410 allocs/op
```

</details>
</details>

4. [Тестовое покрытие для сервиса поиска по XML](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/4)
<details><summary>Result:</summary>

```bash
=== RUN   TestFindUsers
--- PASS: TestFindUsers (0.01s)
=== RUN   TestFindUsersErrorJSON
--- PASS: TestFindUsersErrorJSON (0.00s)
=== RUN   TestFindUsersBrokenResultJSON
--- PASS: TestFindUsersBrokenResultJSON (0.00s)
=== RUN   TestFindUsersFatalError
--- PASS: TestFindUsersFatalError (0.00s)
=== RUN   TestFindUsersTimeOut
--- PASS: TestFindUsersTimeOut (2.00s)
=== RUN   TestFindUsersClientUnknownError
--- PASS: TestFindUsersClientUnknownError (0.00s)
PASS
coverage: 100.0% of statements
ok      hw4     2.864s
```
[HTML-отчет](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/4/cover.html)
</details>

5. [Веб-фреймворк на основе кодогенерации](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/5)
<details><summary>Result:</summary>

```bash
```
</details>

6. [Универсальный сервис просмотра содержимого БД](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/6)
<details><summary>Result:</summary>
```
```
</details>

7. [Асинхронная система логирования](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/7)
<details><summary>Result:</summary>
```
```
</details>

8. [Заполнение полей структуры через рефлексию](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/8)
<details><summary>Result:</summary>
```
```
</details>

9. [Архитектура типового приложения](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/9)
<details><summary>Result:</summary>
```
```
</details>

10. [Телеграм бот](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/10)
<details><summary>Result:</summary>
```
```
</details>

11. [Маркетплейс на основе GraphQL](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/11)
<details><summary>Result:</summary>
```
```
</details>

12. [Многопользовательская MUD на основе асинхрона](https://github.com/ivansevryukov1995/golang_web_services_course/tree/main/12)
<details><summary>Result:</summary>
```
```
</details>



[Вверх](#anchor)
