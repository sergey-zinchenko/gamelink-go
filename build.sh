#!/bin/sh
RELEASE=0.0.1
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build
docker build -t registry.gitlab.smedialink.com/z/gamelink-go .
docker push registry.gitlab.smedialink.com/z/gamelink-go
