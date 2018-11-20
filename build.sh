#!/bin/sh
#PROJECT=registry.gitlab.smedialink.com/z/gamelink-go
PROJECT=gamelink-go
RELEASE=1.0
COMMIT=`git rev-parse --short HEAD`
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X ${PROJECT}/version.Release=${RELEASE} -X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}"
#docker build -t registry.gitlab.smedialink.com/z/gamelink-go .
#docker push registry.gitlab.smedialink.com/z/gamelink-go
