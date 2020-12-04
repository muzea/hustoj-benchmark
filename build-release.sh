#!/bin/bash
set -x
BUILD_DIR=$(dirname "$0")/build
yarn
yarn build
mkdir -p $BUILD_DIR
cd $BUILD_DIR

VERSION=`date -u +%Y%m%d`
LDFLAGS="-X main.VERSION=$VERSION -s -w"
GCFLAGS=""

go get github.com/muzea/hustoj-benchmark

# AMD64
OSES=(linux darwin windows freebsd)
for os in ${OSES[@]}; do
	suffix=""
	if [ "$os" == "windows" ]
	then
		suffix=".exe"
	fi
	env CGO_ENABLED=0 GOOS=$os GOARCH=amd64 go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o timetest_${os}_amd64${suffix} github.com/muzea/hustoj-benchmark
	tar -zcf timetest-${os}-amd64-$VERSION.tar.gz timetest_${os}_amd64${suffix} public
done

# 386
OSES=(linux windows)
for os in ${OSES[@]}; do
	suffix=""
	if [ "$os" == "windows" ]
	then
		suffix=".exe"
	fi
	env CGO_ENABLED=0 GOOS=$os GOARCH=386 go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o timetest_${os}_386${suffix} github.com/muzea/hustoj-benchmark
	tar -zcf timetest-${os}-386-$VERSION.tar.gz timetest_${os}_386${suffix} public
done

# ARM
ARMS=(5 6 7)
for v in ${ARMS[@]}; do
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=$v go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o timetest_linux_arm$v  github.com/muzea/hustoj-benchmark
tar -zcf timetest-linux-arm$v-$VERSION.tar.gz timetest_linux_arm$v public
done

# ARM64
env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o timetest_linux_arm64  github.com/muzea/hustoj-benchmark
tar -zcf timetest-linux-arm64-$VERSION.tar.gz timetest_linux_arm64 public

#MIPS32LE
env CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o timetest_linux_mipsle github.com/muzea/hustoj-benchmark
env CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat go build -ldflags "$LDFLAGS" -gcflags "$GCFLAGS" -o timetest_linux_mips github.com/muzea/hustoj-benchmark

tar -zcf timetest-linux-mipsle-$VERSION.tar.gz timetest_linux_mipsle public
tar -zcf timetest-linux-mips-$VERSION.tar.gz timetest_linux_mips public
