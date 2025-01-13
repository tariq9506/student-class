include .env
# run application server, use command "make run"
run:
	go run main.go

# build for production, use command "make build"
build:
	GOOS=linux GOARCH=amd64 go build -o ./student-apis
	upx student-apis
	tar -czf student-apis.tar.gz student-apis .env.sample 

db_up:
	migrate -path database/ -database "postgresql://$(DBUSER):$(DBPASS)@$(DBHOST):$(DBPORT)/$(DBNAME)?sslmode=disable" -verbose up

db_down:
	migrate -path database/ -database "postgresql://$(DBUSER):$(DBPASS)@$(DBHOST):$(DBPORT)/$(DBNAME)?sslmode=disable" -verbose down $(version)
