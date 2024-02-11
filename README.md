# piju-touchscreen-go

A golang-implementation of a touchscreen UI for piju

Building GTK4+go programs for the Raspberry Pi works better if you cross-compile. If you build on your Pi, (a) you will have to wait a long time, (b) if you have one of the smaller RAM Pis (e.g. a 3 B+) then you may even see build failures due to running out of memory at the linking step.

## Build instructions: on a Pi

Untested, but if you have a Pi with lots of memory, install the relevant version of go, you should just be able to `go mod tidy; go build .`

## Build instructions: Cross-compiling

Cross-compiling relies upon the Docker images at <https://github.com/nsw42/alpine-cross-compile> . See that repo for how to build and tag your Docker images. If you use a non-default tag, you will need to edit the build script (`build.sh`). Similarly, if you're building for a 64-bit OS, you'll need to edit the build script to reference the arm64 builder tag. After that, just run `./build.sh`

## This is too hard!

If the golang version is proving uncooperative, there's a Python based user interface for piju at <https://github.com/nsw42/piju-touchscreen>.

That repo also contains hardware and X setup instructions.
