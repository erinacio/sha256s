#!/bin/sh
VERSION="$(git describe --tags --always)"
export CGO_ENABLED=0
mkdir -p output
build() {
    suffix=''
    if [ "$1" = "windows" ]; then
        suffix='.exe'
    fi
    set -ex
    GOOS=$1 GOARCH=$2 go build -ldflags "-X main.Version=${VERSION}" -o output/sha256s$suffix
    tar -C output -cvzf "output/sha256s.$1.$2.tar.gz" sha256s$suffix
    rm output/sha256s$suffix
}
build darwin amd64
build linux amd64
build windows amd64
# build darwin arm64
build linux arm64
# build windows arm64
build android arm64
