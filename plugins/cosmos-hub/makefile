#!/usr/bin/make -f

plugin:
	go build -buildmode=plugin -o main.so main.go

plugin-debug:
	go build -buildmode=plugin -gcflags="all=-N -l" -o main.so main.go