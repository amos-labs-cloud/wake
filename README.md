## Wake

### Description
Wake is a simple command line utility that allows you to send a WOL magic packet to a device on your network through
pressing a button hooked up to a raspberry pi GPIO pin for signal.

### To package this

You need to have fpm installed, do what you need to do to get that ruby gem installed.

run `make packages`

## Building

run `make build` to build the binary with no cross compilation for where ever you are running the build.
run `make rpi-zero` to build the binary for a raspberry pi zero.
run `make rpi-5` to build the binary for a raspberry pi 5.

