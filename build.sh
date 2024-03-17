#! /bin/bash

docker run -it --rm -v ./:/go/src -v ./dist:/go/output -w /go/src gotk-cross-builder-alpine3.19-armhf sh -c 'go mod tidy; go build -o /go/output/piju-touchscreen-go .'
