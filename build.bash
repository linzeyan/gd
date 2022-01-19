#!/bin/bash

set -ex

VERSION="$(git describe --tags)"
PACKAGE=$(basename ${PWD})
TARGET="bin"

build() {
    if [[ -d ${TARGET} ]]; then
        rm -rf ${TARGET}
    fi
    # Linux
    export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
    go build -a -trimpath -ldflags="-X 'main.appVersion=${VERSION}' -X 'main.appBuildTime=$(date "+%FT%T%Z")' -X 'main.appCommit=$(git rev-parse HEAD)' -X 'main.appPlatform=${GOOS}/${GOARCH}' -X 'main.appGoVersion=$(go version | awk '{print $3}')'" -o ${TARGET}/${PACKAGE}_${VERSION}_${GOOS}_${GOARCH} cmd/*.go
    export GOOS=linux GOARCH=arm64 CGO_ENABLED=0
    go build -a -trimpath -ldflags="-X 'main.appVersion=${VERSION}' -X 'main.appBuildTime=$(date "+%FT%T%Z")' -X 'main.appCommit=$(git rev-parse HEAD)' -X 'main.appPlatform=${GOOS}/${GOARCH}' -X 'main.appGoVersion=$(go version | awk '{print $3}')'" -o ${TARGET}/${PACKAGE}_${VERSION}_${GOOS}_${GOARCH} cmd/*.go
    # Mac
    GOOS=darwin GOARCH=amd64 CGO_ENABLED=0
    go build -a -trimpath -ldflags="-X 'main.appVersion=${VERSION}' -X 'main.appBuildTime=$(date "+%FT%T%Z")' -X 'main.appCommit=$(git rev-parse HEAD)' -X 'main.appPlatform=${GOOS}/${GOARCH}' -X 'main.appGoVersion=$(go version | awk '{print $3}')'" -o ${TARGET}/${PACKAGE}_${VERSION}_${GOOS}_${GOARCH} cmd/*.go
    # Windows
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0
    go build -a -trimpath -ldflags="-X 'main.appVersion=${VERSION}' -X 'main.appBuildTime=$(date "+%FT%T%Z")' -X 'main.appCommit=$(git rev-parse HEAD)' -X 'main.appPlatform=${GOOS}/${GOARCH}' -X 'main.appGoVersion=$(go version | awk '{print $3}')'" -o ${TARGET}/${PACKAGE}_${VERSION}_${GOOS}_${GOARCH} cmd/*.go
}

convert() {
    local compressDir=upx
    if [[ -d ${compressDir} ]]; then
        rm -rf ${compressDir}
        mkdir ${compressDir}
    fi
    for i in $(ls ${TARGET}); do upx -9 -o ${compressDir}/${i} ${TARGET}/${i}; done
    rm -rf ${TARGET}
}

clean() {
    go clean
    rm -rf ${TARGET}
}

createService() {
    cat <<'EOF' >/lib/systemd/system/gd.service
[Unit]
Description=Fetch DNS
After=network.target
After=mysql.service

[Service]
WorkingDirectory=/data/dns
ExecStart=/data/dns/gd -o hourly
ExecReload=/bin/kill -s HUP $MAINPID
Restart=always

[Install]
WantedBy=multi-user.target
EOF
}

$1
