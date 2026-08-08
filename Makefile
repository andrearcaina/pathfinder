.PHONY: default build run vet test check clean

default:
	echo "Code Metric Tool Created By Andre Arcaina"

build:
	go build -o bin/pathfinder main.go
	chmod +x bin/pathfinder

run:
	bash -c "./bin/pathfinder"

vet:
	go vet ./...

test:
	go test ./...

check: vet test build

clean:
	rm -rf ./bin/**
