#! /bin/bash

docker run -it --rm -v ./:/go/src -w /go/src go-gtk-image-linuxarmhf sh -c 'go mod tidy; go build .'
