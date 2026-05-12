test:
	go test -v -race

doc-run:
	docker run -p 3307:3306 -v $(PWD):/docker-entrypoint-initdb.d -e MYSQL_ROOT_PASSWORD=1234 -e MYSQL_DATABASE=golang -d mysql

up:
	docker compose up --build --force-recreate

down:
	docker compose down