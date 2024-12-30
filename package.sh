#!/bin/bash -ex

# Define variables
BUILD_DIR=./build/binaries
PACKAGE_DIR=./packages
BINARY_NAME=wake
SERVICE_FILE=deb/wake.service
CONFIG_FILE=./wake.yaml
VERSION="0.0.1"

# Parse command-line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --version) VERSION="$2"; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

# Create package directory
mkdir -p $PACKAGE_DIR

# Package for Raspberry Pi Zero (armv6)
fpm -s dir -t deb --name $BINARY_NAME --version $VERSION --architecture armhf \
    --description "Wake utility for Raspberry Pi Zero" \
    --config-files /etc/wake/wake.yaml \
    --package $PACKAGE_DIR/${BINARY_NAME}_${VERSION}_armhf.deb \
    $CONFIG_FILE=/etc/wake/wake.yaml \
    $SERVICE_FILE=/usr/lib/systemd/system/wake.service \
    $BUILD_DIR/${BINARY_NAME}-rpi-zero=/usr/local/bin/${BINARY_NAME}

