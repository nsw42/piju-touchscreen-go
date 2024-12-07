# piju-touchscreen-go

A golang-implementation of a touchscreen UI for piju

Building GTK4+go programs for the Raspberry Pi works better if you cross-compile. If you build on your Pi, (a) you will have to wait a long time, (b) if you have one of the smaller RAM Pis (e.g. a 3 B+) then you may even see build failures due to running out of memory at the linking step.

## Build instructions: on a Pi

Untested, but if you have a Pi with lots of memory, install the relevant version of go, you should just be able to `go mod tidy; go build .`

## Build instructions: Cross-compiling

Cross-compiling relies upon the Docker images at <https://github.com/nsw42/alpine-cross-compile> . See that repo for how to build and tag your Docker images. If you use a non-default tag, you will need to edit the build script (`build.sh`). Similarly, if you're building for a 64-bit OS, you'll need to edit the build script to reference the arm64 builder tag. After that, just run `./build.sh`

## Deploying

* scp the compiled binary to the pi
* sudo apk add gtk4.0 util-linux
* As root, edit `/etc/inittab` so that the `tty1` line reads as follows:

  ```text
  tty1::respawn:/sbin/agetty --autologin piju --noclear 38400 tty1
  ```

* If you reboot at this point, you should be automatically logged in (to an ash prompt) as the `piju` user.
* As the `piju` user, create `$HOME/.profile`:

  ```sh
  if [ -z "${DISPLAY}" -a -z "${SSH_CONNECTION}" ]; then
      exec startx
  fi
  ```

* As the `piju` user, create `$HOME/.xinitrc`:

  ```sh
  #! /bin/sh

  exec ./piju-touchscreen-go --host http://SERVER:5000/ --mode dark --layout fixed --fullscreen --screenblanker-profile onoff --hidemousepointer >> /var/log/piju-touchscreen/stdout 2>> /var/log/piju-touchscreen/stderr
  ```

* As the `piju` user, create a directory for the touchscreen UI log files:

  ```sh
  sudo mkdir -m 777 /var/log/piju-touchscreen
  ```

## Known issues

There is a memory leak in the underlying go-gtk library (see <https://github.com/diamondburned/gotk4/issues/126>). The fix for that depends on go 1.24. Until those fixes are available, memory use of the UI steadily increases every time the 'now playing' artwork changes. The simplest fix is to create a cron job that kills the running touchscreen process at a time when people are not likely to be using the UI. If your OS provides `/etc/periodic/daily/` to run such operations, and it also includes `pkill`, add the following shell script as `/etc/periodic/daily/piju-touchscreen-go`:

```sh
#! /bin/sh

pkill -f piju-touchscreen-go
```

If your OS does not provide `/etc/periodic/daily/`, but does provide pkill, instead add the following line to the root crontab by running `sudo crontab -e`:

```text
0 3 * * * pkill -f piju-touchscreen-go
```

## This is too hard!

If the golang version proves uncooperative, there's a Python based user interface for piju at <https://github.com/nsw42/piju-touchscreen>.

That repo also contains hardware and X setup instructions.
