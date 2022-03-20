#!/bin/bash
BUILD_DIR=$(dirname "$0")/build
yarn
yarn build

# AMD64
OSES=(linux darwin windows)
for os in ${OSES[@]}; do
	suffix=""
	if [ "$os" == "windows" ]
	then
		suffix=".exe"
	fi
	env CGO_ENABLED=0 GOOS=$os GOARCH=amd64 go build -o ${BUILD_DIR}/timetest_${os}_amd64${suffix}
	tar -zcf ${BUILD_DIR}/timetest-${os}-amd64.tar.gz -C ${BUILD_DIR} timetest_${os}_amd64${suffix}
done

# 386
OSES=(linux windows)
for os in ${OSES[@]}; do
	suffix=""
	if [ "$os" == "windows" ]
	then
		suffix=".exe"
	fi
	env CGO_ENABLED=0 GOOS=$os GOARCH=386 go build -o ${BUILD_DIR}/timetest_${os}_386${suffix}
	tar -zcf ${BUILD_DIR}/timetest-${os}-386.tar.gz -C ${BUILD_DIR} timetest_${os}_386${suffix}
done

# ARM
env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o ${BUILD_DIR}/timetest_linux_arm7
tar -zcf ${BUILD_DIR}/timetest-linux-arm7.tar.gz -C ${BUILD_DIR} timetest_linux_arm7

# ARM64
env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ${BUILD_DIR}/timetest_linux_arm64
tar -zcf ${BUILD_DIR}/timetest-linux-arm64.tar.gz -C ${BUILD_DIR} timetest_linux_arm64

#MIPS32LE
env CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -o ${BUILD_DIR}/timetest_linux_mipsle
env CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat go build -o ${BUILD_DIR}/timetest_linux_mips

tar -zcf ${BUILD_DIR}/timetest-linux-mipsle.tar.gz -C ${BUILD_DIR} timetest_linux_mipsle
tar -zcf ${BUILD_DIR}/timetest-linux-mips.tar.gz -C ${BUILD_DIR} timetest_linux_mips
