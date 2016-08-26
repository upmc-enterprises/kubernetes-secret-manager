# Makefile for the Docker image stevesloka/kubernetes-secret-manager-sampleapp
# MAINTAINER: Steve Sloka <steve@stevesloka.com>

.PHONY: all build container push clean test

TAG = 1.0.0
PREFIX = stevesloka

all: container

build: main.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o kubernetes-secret-manager-sampleapp --ldflags '-w' ./main.go ./handler.go ./secret-loader.go

container: build
	docker build -t $(PREFIX)/kubernetes-secret-manager-sampleapp:$(TAG) .

push:
	docker push $(PREFIX)/kubernetes-secret-manager-sampleapp:$(TAG)

clean:
	rm -f kubernetes-secret-manager-sampleapp

test: clean
	godep go test -v --vmodule=*=4
