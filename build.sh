#!/bin/sh
#PROJECT=registry.gitlab.smedialink.com/z/gamelink-go
PROJECT=gamelink-go
RELEASE=0.0.1
COMMIT=`git rev-parse --short HEAD`
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go run -ldflags "-X ${PROJECT}.version.Release=${RELEASE} -X ${PROJECT}.version.Commit=${COMMIT} -X ${PROJECT}.version.BuildTime=${BUILD_TIME}" main.go
#docker build -t registry.gitlab.smedialink.com/z/gamelink-go .
#docker push registry.gitlab.smedialink.com/z/gamelink-go
