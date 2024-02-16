.PHONY: tidy dep-update vendor run build

tidy:
	go mod tidy

dep-update:
	go get -u ./...

vendor:
	go mod vendor

run:
	go run cmd/htp/main.go

build:
	go build -o bin/htp cmd/htp/main.go
