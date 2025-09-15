.PHONY: default build run clean

default:
	echo "Code Metric Tool Created By Andre Arcaina"

build:
	go build -o bin/pathfinder main.go
	chmod +x bin/pathfinder

run:
	bash -c "./bin/pathfinder"

clean:
	rm -rf ./bin/**
