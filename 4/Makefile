.PHONY: build
build:
	go build -gcflags="all=-N -l" -o ./bin/server ./client.go

.PHONY: run
run:
	./bin/server

.PHONY: curl-app
curl-app:
	curl -H "AccessToken: test-token" \
		'localhost:8080/?query=&limit=25&offset=0&order_by=1&order_field=Id'

.PHONY: attach
attach:
	dlv attach $$(tasklist | findstr "server" | awk '{print $$2}')

test:
	go test -v -cover

cover:
	go test -v -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html
	@start "" "C:\Users\Ivan\go\src\github.com\ivansevryukov1995\golang_web_services_course\4\cover.html"
	