.PHONY: lint tidy update vendor run build

lint: tidy
	golines -w -m 88 --base-formatter=gofumpt .

tidy:
	go mod tidy

update:
	go get -u ./...

vendor:
	go mod vendor

run:
	go run cmd/htp/main.go

build:
	go build -o bin/htp cmd/htp/main.go
