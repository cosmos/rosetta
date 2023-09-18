#!/usr/bin/make -f

plugin:
	go build -buildmode=plugin -o main.so main.go

docker:
	docker build . --tag rosetta