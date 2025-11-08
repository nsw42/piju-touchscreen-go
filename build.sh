#! /bin/bash

docker run -it --rm -v ./:/go/src -v ./dist:/go/output -w /go/src gotk-cross-builder-armhf-alpine-3.22 sh -c 'go mod tidy; go build -o /go/output/piju-touchscreen-go .'
