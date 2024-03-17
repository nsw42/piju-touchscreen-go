#! /bin/bash

docker run -it --rm -v ./:/go/src -w /go/src gotk-cross-builder-alpine3.19-armhf sh -c 'go mod tidy; go build .'
